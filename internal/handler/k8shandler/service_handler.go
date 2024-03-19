package k8shandler

import (
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	corev1 "k8s.io/api/core/v1"
)

type ServiceHandlerOpts struct{}

type serviceHandler struct {
	clusterID string
}

// NewRemoteClusterSecretHandler creates a new handler for remote cluster secrets
// The handler is responsible to listen for remote cluster secrets
// When a new secret is added/updated, the handler resolves the secret to get the remote cluster config and creates a new remote cluster in cache
// The handler also starts the relevant controllers (Service, rollout, service) for the remote cluster.
func NewServiceHandler(clusterID string, _ ServiceHandlerOpts) handler.Handler {
	serviceHandlerNew := &serviceHandler{
		clusterID: clusterID,
	}

	return serviceHandlerNew
}

func (s *serviceHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	service, ok := obj.(*corev1.Service)
	if !ok {
		ctx.Log.Error("error casting Service object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.ResourceIdentifierKey, service.Name).Str(logger.ClusterKey, s.clusterID).Info("Adding service to cluster service cache")
	cache.Services.Add(s.clusterID, service)
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *serviceHandler) Updated(ctx context.Context, newObj interface{}, _ interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	service, ok := newObj.(*corev1.Service)
	if !ok {
		ctx.Log.Error("error casting Service object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.ResourceIdentifierKey, service.Name).Str(logger.ClusterKey, s.clusterID).Info("Updating service to cluster service cache")
	cache.Services.Add(s.clusterID, service)
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *serviceHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	service, ok := obj.(*corev1.Service)
	if !ok {
		ctx.Log.Error("error casting Service object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.ResourceIdentifierKey, service.Name).Str(logger.ClusterKey, s.clusterID).Info("Deleting service from cluster service cache")
	cache.Services.Delete(s.clusterID, service)
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *serviceHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}
