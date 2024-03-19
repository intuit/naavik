package k8scontroller

import (
	ctx "context"
	"fmt"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type ServiceController struct {
	clientset kubernetes.Interface

	// Informer options
	namespace    string
	resyncPeriod time.Duration
	listOpts     metav1.ListOptions
	handler      handler.Handler
}

func NewServiceController(
	name string,
	k8sConfig *rest.Config,
	configLoader k8s_utils.ClientConfigLoader,
	namespace string,
	listOpts metav1.ListOptions,
	resync time.Duration,
	handler handler.Handler,
) error {
	log := logger.NewLogger()

	controllerName := fmt.Sprintf("service-controller/%s", k8sConfig.Host)
	if len(name) > 0 {
		controllerName = fmt.Sprintf("service-controller/%s", name)
	}

	log.WithStr(logger.ControllerNameKey, controllerName).WithStr(logger.NamespaceKey, namespace).WithStr(logger.LabelSelectorKey, listOpts.LabelSelector).Info("Initializing controller")

	client, err := configLoader.ClientFromConfig(k8sConfig)
	if err != nil {
		log.WithStr(logger.ErrorKey, err.Error()).Error("error creating k8s client from config")
		return err
	}

	controller.NewController(controller.Opts{
		Name: controllerName,
		Delegator: &ServiceController{
			clientset:    client,
			namespace:    namespace,
			listOpts:     listOpts,
			handler:      handler,
			resyncPeriod: resync,
		},
	})
	return nil
}

func (s *ServiceController) GetInformer() cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return s.clientset.CoreV1().Services(s.namespace).List(ctx.Background(), s.listOpts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return s.clientset.CoreV1().Services(s.namespace).Watch(ctx.Background(), s.listOpts)
			},
		},
		&corev1.Service{}, s.resyncPeriod, cache.Indexers{},
	)
}

func (s *ServiceController) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	s.handler.Added(ctx, obj, statusChan)
}

func (s *ServiceController) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) {
	s.handler.Updated(ctx, newObj, oldObj, statusChan)
}

func (s *ServiceController) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	s.handler.Deleted(ctx, obj, statusChan)
}

func (s *ServiceController) Status(ctx context.Context, eventStatus controller.EventProcessStatus) {
	s.handler.OnStatus(ctx, eventStatus)
}
