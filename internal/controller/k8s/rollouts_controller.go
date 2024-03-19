package k8scontroller

import (
	"fmt"
	"time"

	argoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	argoinformers "github.com/argoproj/argo-rollouts/pkg/client/informers/externalversions"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type RolloutsController struct {
	k8sConfig *rest.Config
	clientset argoclientset.Interface

	// Informer options
	namespace    string
	resyncPeriod time.Duration
	listOpts     metav1.ListOptions
	handler      handler.Handler
}

func NewRolloutsController(
	name string,
	k8sConfig *rest.Config,
	configLoader k8s_utils.ClientConfigLoader,
	namespace string,
	listOpts metav1.ListOptions,
	resync time.Duration,
	handler handler.Handler,
) error {
	log := logger.NewLogger()

	controllerName := fmt.Sprintf("rollouts-controller/%s", k8sConfig.Host)
	if len(name) > 0 {
		controllerName = fmt.Sprintf("rollouts-controller/%s", name)
	}

	log.WithStr(logger.ControllerNameKey, controllerName).WithStr(logger.NamespaceKey, namespace).WithStr(logger.LabelSelectorKey, listOpts.LabelSelector).Info("Initializing controller")

	client, err := configLoader.ArgoClientFromConfig(k8sConfig)
	if err != nil {
		log.WithStr(logger.ErrorKey, err.Error()).Error("error creating argo client from config")
		return err
	}

	controller.NewController(controller.Opts{
		Name: controllerName,
		Delegator: &RolloutsController{
			clientset:    client,
			namespace:    namespace,
			listOpts:     listOpts,
			handler:      handler,
			resyncPeriod: resync,
		},
	})

	return nil
}

func (rc *RolloutsController) GetInformer() cache.SharedIndexInformer {
	argoRolloutsInformerFactory := argoinformers.NewSharedInformerFactoryWithOptions(
		rc.clientset,
		rc.resyncPeriod,
		argoinformers.WithNamespace(metav1.NamespaceAll))
	return argoRolloutsInformerFactory.Argoproj().V1alpha1().Rollouts().Informer()
}

func (rc *RolloutsController) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	rc.handler.Added(ctx, obj, statusChan)
}

func (rc *RolloutsController) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) {
	rc.handler.Updated(ctx, newObj, oldObj, statusChan)
}

func (rc *RolloutsController) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	rc.handler.Deleted(ctx, obj, statusChan)
}

func (rc *RolloutsController) Status(ctx context.Context, eventStatus controller.EventProcessStatus) {
	rc.handler.OnStatus(ctx, eventStatus)
}
