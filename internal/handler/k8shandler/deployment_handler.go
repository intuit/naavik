package k8shandler

import (
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	traffic_config "github.com/intuit/naavik/internal/handler/trafficconfig"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	v1 "k8s.io/api/apps/v1"
)

type DeploymentHandlerOpts struct{}

type deploymentHandler struct {
	clusterID string
}

// NewRemoteClusterSecretHandler creates a new handler for remote cluster secrets
// The handler is responsible to listen for remote cluster secrets
// When a new secret is added/updated, the handler resolves the secret to get the remote cluster config and creates a new remote cluster in cache
// The handler also starts the relevant controllers (deployment, rollout, service) for the remote cluster.
func NewDeploymentHandler(clusterID string, _ DeploymentHandlerOpts) handler.Handler {
	deploymentHandler := &deploymentHandler{
		clusterID: clusterID,
	}

	return deploymentHandler
}

func (d *deploymentHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	deploy, ok := obj.(*v1.Deployment)
	if !ok {
		ctx.Log.Error("error casting deployment object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	// Ignore the deployment if it is not mesh enabled
	if utils.ResourceUtil().IsResourceIgnored(deploy.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(deploy.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	workloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(deploy.Spec.Template.ObjectMeta)
	if len(workloadIdentifier) == 0 {
		ctx.Log.Error("Deployment workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.WorkloadIdentifierKey, workloadIdentifier).Str(logger.ClusterKey, d.clusterID).Info("Adding deploy to cluster deployment cache")
	cache.IdentityCluster.AddClusterToIdentity(workloadIdentifier, d.clusterID)
	cache.Deployments.Add(d.clusterID, deploy)

	if !options.IsCacheWarmedUp() {
		ctx.Log.Info("Cache not warmed up, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	tcHandler := traffic_config.NewTrafficConfigHandler()
	tcHandler.TriggerTrafficConfigHandlerForIdentity(ctx, workloadIdentifier, statusChan)

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (d *deploymentHandler) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	newDeploy, newok := newObj.(*v1.Deployment)
	oldDeploy, oldok := oldObj.(*v1.Deployment)
	if !newok || !oldok {
		ctx.Log.Error("error casting deployment object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Ignore the deployment if it is not mesh enabled
	if utils.ResourceUtil().IsResourceIgnored(newDeploy.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(newDeploy.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// TODO: Should we only take the mesh enabled deployments?
	newWrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(newDeploy.Spec.Template.ObjectMeta)
	oldWrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(oldDeploy.Spec.Template.ObjectMeta)

	if len(newWrkloadIdentifier) == 0 {
		ctx.Log.Error("Deployment workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	if len(oldWrkloadIdentifier) > 0 && newWrkloadIdentifier != oldWrkloadIdentifier {
		ctx.Log.Str(logger.WorkloadIdentifierKey, newWrkloadIdentifier).Str(fmt.Sprintf("old_%s", logger.WorkloadIdentifierKey), oldWrkloadIdentifier).Str(logger.ClusterKey, d.clusterID).Info("Deployment workload identifier changed, adding cluster to new identity")
		cache.IdentityCluster.DeleteClusterFromIdentity(oldWrkloadIdentifier, d.clusterID)
		cache.Deployments.Delete(d.clusterID, oldDeploy)
	} else if len(newWrkloadIdentifier) > 0 {
		ctx.Log.Str(logger.WorkloadIdentifierKey, newWrkloadIdentifier).Str(logger.ClusterKey, d.clusterID).Info("Deployment workload identifier did not change, adding identity to cluster cache")
	}
	cache.IdentityCluster.AddClusterToIdentity(newWrkloadIdentifier, d.clusterID)
	cache.Deployments.Add(d.clusterID, newDeploy)

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (d *deploymentHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	deploy, ok := obj.(*v1.Deployment)
	if !ok {
		ctx.Log.Error("error casting deployment object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	// Ignore the deployment if it is not mesh enabled or is ignored
	if utils.ResourceUtil().IsResourceIgnored(deploy.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(deploy.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	wrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(deploy.Spec.Template.ObjectMeta)
	if len(wrkloadIdentifier) == 0 {
		ctx.Log.Error("Deployment workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	ctx.Log.Str(logger.WorkloadIdentifierKey, wrkloadIdentifier).Str(logger.ClusterKey, d.clusterID).Trace("Deployment workload identifier deleted from deployments and cluster cache")

	cache.Deployments.Delete(d.clusterID, deploy)
	cache.IdentityCluster.DeleteClusterFromIdentity(wrkloadIdentifier, d.clusterID)
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (d *deploymentHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}
