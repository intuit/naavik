package admiral

import (
	"fmt"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	clientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	admiralinformerv1alpha1 "github.com/istio-ecosystem/admiral-api/pkg/client/informers/externalversions/admiral/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const DependencyControllerNameValue = "dependency-controller"

type dependencyController struct {
	k8sConfig *rest.Config
	clientset clientset.Interface
	// Informer options
	namespace    string
	resyncPeriod time.Duration
	listOpts     metav1.ListOptions
	handler      handler.Handler
}

func NewDependencyController(
	name string,
	k8sConfig *rest.Config,
	configLoader k8s_utils.ClientConfigLoader,
	namespace string,
	listOpts metav1.ListOptions,
	resync time.Duration,
	handler handler.Handler,
) error {
	log := logger.NewLogger()

	controllerName := fmt.Sprintf("%s/%s", DependencyControllerNameValue, k8sConfig.Host)
	if len(name) > 0 {
		controllerName = fmt.Sprintf("%s/%s", DependencyControllerNameValue, name)
	}

	log.WithStr(logger.ControllerNameKey, controllerName).WithStr(logger.NamespaceKey, namespace).WithStr(logger.LabelSelectorKey, listOpts.LabelSelector).Info("Initializing controller")

	client, err := configLoader.AdmiralClientFromConfig(k8sConfig)
	if err != nil {
		log.WithStr(logger.ErrorKey, err.Error()).Error("error creating k8s client from config")
		return err
	}

	controller.NewController(controller.Opts{
		Name: controllerName,
		Delegator: &dependencyController{
			clientset:    client,
			namespace:    namespace,
			listOpts:     listOpts,
			handler:      handler,
			resyncPeriod: resync,
		},
	})
	return nil
}

func (dc *dependencyController) GetInformer() cache.SharedIndexInformer {
	return admiralinformerv1alpha1.NewDependencyInformer(
		dc.clientset,
		dc.namespace,
		dc.resyncPeriod,
		cache.Indexers{},
	)
}

func (dc *dependencyController) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	dc.handler.Added(ctx, obj, statusChan)
}

func (dc *dependencyController) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) {
	dc.handler.Updated(ctx, newObj, oldObj, statusChan)
}

func (dc *dependencyController) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	dc.handler.Deleted(ctx, obj, statusChan)
}

func (dc *dependencyController) Status(ctx context.Context, eventStatus controller.EventProcessStatus) {
	dc.handler.OnStatus(ctx, eventStatus)
}
