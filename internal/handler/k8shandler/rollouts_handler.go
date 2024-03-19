package k8shandler

import (
	"fmt"

	argo "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	traffic_config "github.com/intuit/naavik/internal/handler/trafficconfig"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
)

type RolloutHandlerOpts struct{}

type rolloutHandler struct {
	clusterID string
}

// NewRemoteClusterSecretHandler creates a new handler for remote cluster secrets
// The handler is responsible to listen for remote cluster secrets
// When a new secret is added/updated, the handler resolves the secret to get the remote cluster config and creates a new remote cluster in cache
// The handler also starts the relevant controllers (Rollout, rollout, service) for the remote cluster.
func NewRolloutHandler(clusterID string, _ RolloutHandlerOpts) handler.Handler {
	rolloutHandlerNew := &rolloutHandler{
		clusterID: clusterID,
	}

	return rolloutHandlerNew
}

func (r *rolloutHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	rollout, ok := obj.(*argo.Rollout)
	if !ok {
		ctx.Log.Error("error casting Rollout object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	// Ignore the rollout if it is not mesh enabled or is ignored
	if utils.ResourceUtil().IsResourceIgnored(rollout.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(rollout.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	workloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(rollout.Spec.Template.ObjectMeta)
	if len(workloadIdentifier) == 0 {
		ctx.Log.Error("Rollout workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	ctx.Log.Str(logger.WorkloadIdentifierKey, workloadIdentifier).Str(logger.ClusterKey, r.clusterID).Info("Adding rollout to cluster rollout cache")
	cache.IdentityCluster.AddClusterToIdentity(workloadIdentifier, r.clusterID)
	cache.Rollouts.Add(r.clusterID, rollout)

	if !options.IsCacheWarmedUp() {
		ctx.Log.Info("Cache not warmed up, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	tcHandler := traffic_config.NewTrafficConfigHandler()
	tcHandler.TriggerTrafficConfigHandlerForIdentity(ctx, workloadIdentifier, statusChan)

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (r *rolloutHandler) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	oldRollout, oldok := oldObj.(*argo.Rollout)
	newRollout, newok := newObj.(*argo.Rollout)
	if !oldok || !newok {
		ctx.Log.Error("error casting Rollout object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Ignore the rollout if it is not mesh enabled
	if utils.ResourceUtil().IsResourceIgnored(newRollout.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(newRollout.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// TODO: Should we only take the mesh enabled rollouts?
	newWrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(newRollout.Spec.Template.ObjectMeta)
	oldWrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(oldRollout.Spec.Template.ObjectMeta)

	if len(newWrkloadIdentifier) == 0 {
		ctx.Log.Error("Rollout workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	if len(oldWrkloadIdentifier) > 0 && newWrkloadIdentifier != oldWrkloadIdentifier {
		ctx.Log.Str(logger.WorkloadIdentifierKey, newWrkloadIdentifier).Str(logger.ClusterKey, r.clusterID).Str(fmt.Sprintf("old_%s", logger.WorkloadIdentifierKey), oldWrkloadIdentifier).Info("Rollout workload identifier changed, adding cluster to new identity")
		cache.IdentityCluster.DeleteClusterFromIdentity(oldWrkloadIdentifier, r.clusterID)
		cache.Rollouts.Delete(r.clusterID, oldRollout)
	} else if len(newWrkloadIdentifier) > 0 {
		ctx.Log.Str(logger.WorkloadIdentifierKey, newWrkloadIdentifier).Str(logger.ClusterKey, r.clusterID).Info("Rollout workload identifier did not change, adding identity to cluster cache")
	}
	cache.IdentityCluster.AddClusterToIdentity(newWrkloadIdentifier, r.clusterID)
	cache.Rollouts.Add(r.clusterID, newRollout)

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (r *rolloutHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	rollout, ok := obj.(*argo.Rollout)
	if !ok {
		ctx.Log.Error("error casting Rollout object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Ignore the rollout if it is not mesh enabled
	if utils.ResourceUtil().IsResourceIgnored(rollout.Spec.Template.ObjectMeta) || !utils.ResourceUtil().IsResourceMeshEnabled(rollout.Spec.Template.ObjectMeta) {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	wrkloadIdentifier := utils.ResourceUtil().GetWorkloadIdentifier(rollout.Spec.Template.ObjectMeta)
	if len(wrkloadIdentifier) == 0 {
		ctx.Log.Error("Rollout workload identifier is empty, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.WorkloadIdentifierKey, wrkloadIdentifier).Str(logger.ClusterKey, r.clusterID).Info("Rollout workload identifier deleted from rollouts and cluster cache")

	cache.Rollouts.Delete(r.clusterID, rollout)
	cache.IdentityCluster.DeleteClusterFromIdentity(wrkloadIdentifier, r.clusterID)
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (r *rolloutHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}
