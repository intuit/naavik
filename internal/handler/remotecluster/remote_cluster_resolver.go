package remotecluster

import (
	"fmt"
	"strings"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller/admiral"
	k8s_controller "github.com/intuit/naavik/internal/controller/k8s"
	k8s_handlers "github.com/intuit/naavik/internal/handler/k8shandler"
	"github.com/intuit/naavik/internal/handler/remotecluster/resolver"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/internal/types/remotecluster"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Resolver interface {
	GetConfigLoader() k8s_utils.ClientConfigLoader
	ResolveConfig(ctx context.Context, secretKey, secretIdentifier string, secretValue []byte) (remotecluster.RemoteCluster, error)
	StartRemoteClusterControllers(ctx context.Context, cluster remotecluster.RemoteCluster)
	StopRemoteClusterControllers(ctx context.Context, clusterID string)
	AddRemoteClusterToCache(ctx context.Context, remoteCluster remotecluster.RemoteCluster)
	DeleteRemoteClusterFromCache(ctx context.Context, clusterID string)
	GetRemoteClusterFromCache(ctx context.Context, clusterID string) (remotecluster.RemoteCluster, bool)
}

type remoteClusterResolver struct {
	configResolver     resolver.ConfigResolver
	clientConfigLoader k8s_utils.ClientConfigLoader
}

// NewRemoteClusterResolver is resolves the secret into a k8s config and client
// It also implements methods to interact with the remote cluster cache.
func NewRemoteClusterResolver(configResolver resolver.ConfigResolver, clientConfigLoader k8s_utils.ClientConfigLoader) Resolver {
	return &remoteClusterResolver{
		configResolver:     configResolver,
		clientConfigLoader: clientConfigLoader,
	}
}

func (rcr *remoteClusterResolver) GetConfigLoader() k8s_utils.ClientConfigLoader {
	return rcr.clientConfigLoader
}

func (rcr *remoteClusterResolver) ResolveConfig(ctx context.Context, secretKey, secretIdentifier string, secretValue []byte) (remotecluster.RemoteCluster, error) {
	startTime := time.Now()
	ctx.Log.Str(logger.NameKey, secretKey).Infof("Resolving k8s config for secret.")
	// Get the kubeconfig from the K8s config resolver with the secret key
	kubeConfig, err := rcr.configResolver.GetKubeConfig(secretKey, secretValue)
	if err != nil {
		return nil, fmt.Errorf("error resolving kubeconfig for secret %s: %v", secretKey, err)
	}

	// Load the kubeconfig to get the cluster config
	clusterConfig, err := rcr.clientConfigLoader.Load(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig for secret %s: %v", secretKey, err)
	}

	clientConfig := rcr.clientConfigLoader.DefaultClientConfig(*clusterConfig, &clientcmd.ConfigOverrides{})

	config, err := rcr.GetConfigLoader().RawConfigFromClientConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting raw config from client config: %v", err)
	}

	restConfig, err := rcr.clientConfigLoader.ClientConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating rest config from client config for secret %s: %v", secretKey, err)
	}

	k8sClient, err := rcr.clientConfigLoader.ClientFromConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client from rest config for secret %s: %v", secretKey, err)
	}

	istioclient, err := rcr.GetConfigLoader().IstioClientFromConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating istio client from config: %v", err)
	}

	argoClient, err := rcr.GetConfigLoader().ArgoClientFromConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating argo client from config: %v", err)
	}

	admiralClient, err := rcr.GetConfigLoader().AdmiralClientFromConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating admiral client from config: %v", err)
	}

	ctx.Log.Str(logger.NameKey, secretKey).Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Infof("Resolved k8s config for secret.")
	cluster := remotecluster.CreateRemoteCluster(secretKey, config.CurrentContext, secretIdentifier,
		restConfig.Host, clientConfig, k8sClient, istioclient, argoClient, admiralClient)

	return cluster, nil
}

func (rcr *remoteClusterResolver) StartRemoteClusterControllers(ctx context.Context, cluster remotecluster.RemoteCluster) {
	ctx.Log.Infof("Starting remote controllers.")

	clientConfig, err := cluster.GetClientConfig().ClientConfig()
	if err != nil {
		ctx.Log.Errorf("error getting client config for remote cluster: %v", err)
		return
	}

	rawConfig, err := cluster.GetClientConfig().RawConfig()
	if err != nil {
		ctx.Log.Errorf("error getting raw config for remote cluster: %v", err)
		return
	}

	k8s_controller.NewServiceController(
		rawConfig.CurrentContext,
		clientConfig,
		rcr.clientConfigLoader,
		corev1.NamespaceAll,
		metav1.ListOptions{},
		options.GetCacheRefreshInterval(),
		k8s_handlers.NewServiceHandler(cluster.GetClusterID(), k8s_handlers.ServiceHandlerOpts{}),
	)

	k8s_controller.NewDeploymentController(
		rawConfig.CurrentContext,
		clientConfig,
		rcr.clientConfigLoader,
		corev1.NamespaceAll,
		metav1.ListOptions{},
		options.GetCacheRefreshInterval(),
		k8s_handlers.NewDeploymentHandler(cluster.GetClusterID(), k8s_handlers.DeploymentHandlerOpts{}),
	)

	if options.IsArgoRolloutsEnabled() {
		ctx.Log.Infof("Starting rollout controller for remote cluster.")
		k8s_controller.NewRolloutsController(
			rawConfig.CurrentContext,
			clientConfig,
			rcr.clientConfigLoader,
			corev1.NamespaceAll,
			metav1.ListOptions{},
			options.GetCacheRefreshInterval(),
			k8s_handlers.NewRolloutHandler(cluster.GetClusterID(), k8s_handlers.RolloutHandlerOpts{}),
		)
	}
}

func (rcr *remoteClusterResolver) StopRemoteClusterControllers(ctx context.Context, clusterID string) {
	ctx.Log.Infof("Stopping remote controllers.")
	cache.ControllerCache.Range(func(key, value interface{}) bool {
		controllerCtx, ok := value.(cache.ControllerContext)
		if ok {
			controllerName := key.(string)
			controllerNameSplit := strings.SplitN(controllerName, "/", 2)
			// Sanity check to make sure we are only stopping the controllers for the cluster
			// and not stopping the cluster secret controller
			if len(controllerNameSplit) == 2 && strings.Contains(controllerNameSplit[1], clusterID) &&
				k8s_controller.SecretControllerNameValue != controllerNameSplit[0] && admiral.TrafficConfigControllerNameValue != controllerNameSplit[0] &&
				admiral.DependencyControllerNameValue != controllerNameSplit[0] {
				go func(ctx context.Context, controller cache.ControllerContext) {
					startTime := time.Now()
					cache.ControllerCache.DeRegister(controllerName)
					ctx.Log.Str(logger.ControllerNameKey, controllerName).Infof("Stopping controller")
					close(controllerCtx.StopCh)
					for i, closeCtx := range controllerCtx.WorkerCtx {
						ctx.Log.Infof("Waiting for workers to finish... %d/%d", i+1, len(controllerCtx.WorkerCtx))
						<-closeCtx.Done()
					}
					// Remove the cluster from the cache only after workers have finished
					cache.RemoteCluster.DeleteCluster(clusterID)
					ctx.Log.Str(logger.ControllerNameKey, controllerName).Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Infof("Stopped controller")
				}(ctx, controllerCtx)
			}
		}
		return true
	})
}

func (rcr *remoteClusterResolver) AddRemoteClusterToCache(ctx context.Context, remoteCluster remotecluster.RemoteCluster) {
	ctx.Log.Str(logger.ClusterKey, remoteCluster.GetClusterID()).Infof("Adding remote cluster to remote cluster cache.")
	cache.RemoteCluster.AddCluster(remoteCluster)
}

func (rcr *remoteClusterResolver) GetRemoteClusterFromCache(_ context.Context, clusterID string) (remotecluster.RemoteCluster, bool) {
	return cache.RemoteCluster.GetCluster(clusterID)
}

func (rcr *remoteClusterResolver) DeleteRemoteClusterFromCache(ctx context.Context, clusterID string) {
	ctx.Log.Str(logger.ClusterKey, clusterID).Infof("Deleting remote cluster from remote cluster cache.")
	cache.RemoteCluster.DeleteCluster(clusterID)
}
