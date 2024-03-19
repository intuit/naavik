package controller

import (
	"fmt"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	"github.com/intuit/naavik/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type FakeController struct {
	Clientset kubernetes.Interface
	Name      string
	// Informer options
	Namespace    string
	ResyncPeriod time.Duration
	ListOpts     metav1.ListOptions
	Handler      handler.Handler
	Informer     cache.SharedIndexInformer
}

// Creates a new mock controller that can be used for testing
// This is a mock controller on deployments that uses mock handler.
func NewFakeController(
	ctx context.Context,
	name string,
	k8sConfig *rest.Config,
	configLoader k8s_utils.ClientConfigLoader,
	namespace string,
	listOpts metav1.ListOptions,
	_ time.Duration,
	handler handler.Handler,
) *FakeController {
	log := logger.NewLogger()

	controllerName := fmt.Sprintf("mock-controller/%s", k8sConfig.Host)
	if len(name) > 0 {
		controllerName = fmt.Sprintf("mock-controller/%s", name)
	}

	log.WithStr(logger.ControllerNameKey, controllerName).WithStr(logger.NamespaceKey, namespace).WithStr(logger.LabelSelectorKey, listOpts.LabelSelector).Info("Initializing controller")

	client, err := configLoader.ClientFromConfig(k8sConfig)
	if err != nil {
		log.WithStr(logger.ErrorKey, err.Error()).Error("error creating k8s client from config")
		return nil
	}

	fakeController := &FakeController{
		Name:      controllerName,
		Clientset: client,
		Namespace: namespace,
		ListOpts:  listOpts,
		Handler:   handler,
	}
	controller.NewController(controller.Opts{
		Context:   ctx,
		Name:      controllerName,
		Delegator: fakeController,
	})
	return fakeController
}

func (s *FakeController) GetInformer() cache.SharedIndexInformer {
	if s.Informer != nil {
		return s.Informer
	}
	s.Informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return s.Clientset.AppsV1().Deployments(s.Namespace).List(context.Background(), s.ListOpts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return s.Clientset.AppsV1().Deployments(s.Namespace).Watch(context.Background(), s.ListOpts)
			},
		},
		&appsv1.Deployment{}, s.ResyncPeriod, cache.Indexers{},
	)
	return s.Informer
}

func (s *FakeController) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	s.Handler.Added(ctx, obj, statusChan)
}

func (s *FakeController) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) {
	s.Handler.Updated(ctx, newObj, oldObj, statusChan)
}

func (s *FakeController) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	s.Handler.Deleted(ctx, obj, statusChan)
}

func (s *FakeController) Status(ctx context.Context, eventStatus controller.EventProcessStatus) {
	s.Handler.OnStatus(ctx, eventStatus)
}
