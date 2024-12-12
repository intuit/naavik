package bootstrap

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/intuit/naavik/cmd/options"
	admiral_controller "github.com/intuit/naavik/internal/controller/admiral"
	k8s_controller "github.com/intuit/naavik/internal/controller/k8s"
	dependency_handler "github.com/intuit/naavik/internal/handler/dependency"
	"github.com/intuit/naavik/internal/handler/remotecluster"
	"github.com/intuit/naavik/internal/handler/remotecluster/resolver"
	trafficconfig_handler "github.com/intuit/naavik/internal/handler/trafficconfig"
	"github.com/intuit/naavik/internal/types/context"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

var configLoader = k8s_utils.NewConfigLoader()

func StartControllers(ctx context.Context) {
	klog.SetLogFilter(newKlogFilter())

	// Get the k8s client config of the cluster where remote cluster dependency, traffic config and remote cluster secrets are stored
	k8sConfig, err := configLoader.GetConfigFromPath(options.GetKubeConfigPath())
	if err != nil {
		log.Fatal(fmt.Sprintf("error getting k8s config from path: %v", err))
	}

	// Initialize controllers
	startSecretController(ctx, k8sConfig.Host, k8sConfig)
	startWorkloadDependencyController(ctx, k8sConfig.Host, k8sConfig)
	startTrafficConfigController(ctx, k8sConfig.Host, k8sConfig)
}

// Secret controller is used to watch for secrets that contain remote k8s config.
func startSecretController(ctx context.Context, _ string, k8sConfig *rest.Config) {
	listOpts := metav1.ListOptions{LabelSelector: options.GetSecretSyncLabel() + "=true"}
	namespace := options.GetClusterRegistriesNamespace()
	resyncPeriod := options.GetCacheRefreshInterval()

	k8sConfigResolver := resolver.GetConfigResolver(ctx, options.GetConfigResolver())
	k8s_controller.NewSecretController(k8sConfig.Host, k8sConfig, configLoader, namespace, listOpts, resyncPeriod,
		remotecluster.NewRemoteClusterSecretHandler(remotecluster.SecretHandlerOpts{
			RemoteClusterResolver: remotecluster.NewRemoteClusterResolver(k8sConfigResolver, configLoader),
		}))
}

func startWorkloadDependencyController(ctx context.Context, name string, k8sConfig *rest.Config) {
	listOpts := metav1.ListOptions{}
	namespace := options.GetDependenciesNamespace()
	resyncPeriod := options.GetCacheRefreshInterval()
	ctx.Log.Str(logger.NameKey, name).Int("resyncPeriod", int(resyncPeriod.Milliseconds())).Info("Initializing workload dependency controller")

	admiral_controller.NewDependencyController(k8sConfig.Host, k8sConfig, configLoader, namespace, listOpts, resyncPeriod,
		dependency_handler.NewDependencyHandler(dependency_handler.Opts{}),
	)
}

func startTrafficConfigController(ctx context.Context, name string, k8sConfig *rest.Config) {
	listOpts := metav1.ListOptions{}
	namespace := options.GetTrafficConfigNamespace()
	// Do not resync, only watch for changes
	resyncPeriod := 0 * time.Second
	ctx.Log.Str(logger.NameKey, name).Int("resyncPeriod", int(resyncPeriod.Milliseconds())).Info("Initializing traffic config controller")

	admiral_controller.NewTrafficConfigController(k8sConfig.Host, k8sConfig, configLoader, namespace, listOpts, resyncPeriod,
		trafficconfig_handler.NewTrafficConfigHandler(),
	)
}

// Wrapper around klog.LogFilter to prevent klog from logging errors too frequently in different format.
type rudimentaryErrorBackoff struct {
	minPeriod time.Duration // immutable
	// TODO(lavalamp): use the clock for testability. Need to move that
	// package for that to be accessible here.
	lastErrorTimeLock sync.Mutex
	lastErrorTime     time.Time
}

// OnError will block if it is called more often than the embedded period time.
// This will prevent overly tight hot error loops.
func (r *rudimentaryErrorBackoff) OnError(error) {
	now := time.Now() // start the timer before acquiring the lock
	r.lastErrorTimeLock.Lock()
	d := now.Sub(r.lastErrorTime)
	r.lastErrorTime = time.Now()
	r.lastErrorTimeLock.Unlock()

	// Do not sleep with the lock held because that causes all callers of HandleError to block.
	// We only want the current goroutine to block.
	// A negative or zero duration causes time.Sleep to return immediately.
	// If the time moves backwards for any reason, do nothing.
	time.Sleep(r.minPeriod - d)
}

type klogFilter struct{}

func newKlogFilter() klog.LogFilter {
	return &klogFilter{}
}

func (*klogFilter) Filter(args []interface{}) []interface{} {
	// No Op
	return args
}

func (*klogFilter) FilterF(format string, args []interface{}) (string, []interface{}) {
	// Log with custom logger
	if strings.Contains(format, "failed to list") {
		logger.Log.Warnf(format, args...)
	}
	return "", nil
}

func (*klogFilter) FilterS(msg string, keysAndValues []interface{}) (string, []interface{}) {
	// No Op
	return msg, keysAndValues
}
