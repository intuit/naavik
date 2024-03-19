package remotecluster

import (
	"fmt"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

type SecretHandlerOpts struct {
	RemoteClusterResolver Resolver
}

type SecretHandler struct {
	RemoteClusterResolver Resolver
}

// NewRemoteClusterSecretHandler creates a new handler for remote cluster secrets
// The handler is responsible to listen for remote cluster secrets
// When a new secret is added/updated, the handler resolves the secret to get the remote cluster config and creates a new remote cluster in cache
// The handler also starts the relevant controllers (deployment, rollout, service) for the remote cluster.
func NewRemoteClusterSecretHandler(opts SecretHandlerOpts) handler.Handler {
	remoteClusterSecretHandler := &SecretHandler{
		RemoteClusterResolver: opts.RemoteClusterResolver,
	}
	return remoteClusterSecretHandler
}

func (s *SecretHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		ctx.Log.Error("error casting secret object, skipping handling of cluster secret.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	if !utils.ResourceUtil().IsSyncEnabled(secret.ObjectMeta) {
		ctx.Log.Info("Resource is ignored, skipping handling of cluster secret.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	for clusterID, kubeConfig := range secret.Data {
		secretIdentifier := getSecretIdentifier(secret.Namespace, secret.Name)
		ctx.Log.WithStr(logger.ClusterKey, clusterID).WithStr(logger.NamespaceSecretKey, secretIdentifier)

		cluster, ok := s.RemoteClusterResolver.GetRemoteClusterFromCache(ctx, clusterID)
		if !ok {
			cluster, err := s.RemoteClusterResolver.ResolveConfig(ctx, clusterID, secretIdentifier, kubeConfig)
			if err != nil {
				ctx.Log.Errorf("unable to resolve k8s config for secret. %s", err.Error())
				continue
			}

			s.RemoteClusterResolver.AddRemoteClusterToCache(ctx, cluster)
			ctx.Log.Info("Cluster added to cache, starting remote cluster controllers.")
			s.RemoteClusterResolver.StartRemoteClusterControllers(ctx, cluster)
		} else if cluster.GetSecretIdentifier() != secretIdentifier {
			ctx.Log.Errorf("cluster already exists in cache with secret %s. Cannot add cluster with secret", cluster.GetSecretIdentifier())
		} else {
			// TODO: Handle update of secret value
			ctx.Log.Warn("cluster already exists in cache with secret. Skipping.")
		}
	}

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *SecretHandler) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	newSecret, ok := newObj.(*corev1.Secret)
	if !ok {
		ctx.Log.Error("error casting new secret object, skipping handling of cluster secret.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	oldSecret, ok := oldObj.(*corev1.Secret)
	if !ok {
		ctx.Log.Error("error casting old secret object, skipping handling of cluster secret.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Stop remote cluster controllers and delete remote cluster from cache if sync is disabled
	if utils.ResourceUtil().IsSyncEnabled(oldSecret.ObjectMeta) && !utils.ResourceUtil().IsSyncEnabled(newSecret.ObjectMeta) {
		for clusterID := range newSecret.Data {
			secretIdentifier := getSecretIdentifier(newSecret.Namespace, newSecret.Name)
			ctx.Log.WithStr(logger.ClusterKey, clusterID).WithStr(logger.NamespaceSecretKey, secretIdentifier)
			ctx.Log.Info("Stopping remote cluster controllers and deleting remote cluster from cache")

			s.RemoteClusterResolver.StopRemoteClusterControllers(ctx, clusterID)
			s.RemoteClusterResolver.DeleteRemoteClusterFromCache(ctx, clusterID)
		}
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	return s.Added(ctx, newObj, statusChan)
}

func (s *SecretHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		ctx.Log.Error("error casting secret object, skipping handling of cluster secret.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	for clusterID := range secret.Data {
		secretIdentifier := getSecretIdentifier(secret.Namespace, secret.Name)
		ctx.Log.WithStr(logger.ClusterKey, clusterID).WithStr(logger.NamespaceSecretKey, secretIdentifier)
		ctx.Log.Info("Stopping remote cluster controllers and deleting remote cluster from cache")

		s.RemoteClusterResolver.StopRemoteClusterControllers(ctx, clusterID)
		s.RemoteClusterResolver.DeleteRemoteClusterFromCache(ctx, clusterID)
	}
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *SecretHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}

func getSecretIdentifier(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}
