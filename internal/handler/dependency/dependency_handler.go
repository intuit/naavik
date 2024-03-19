package dependencyhandler

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	traffic_config "github.com/intuit/naavik/internal/handler/trafficconfig"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	admiralApi "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
)

type Opts struct{}

type dependencyHandler struct{}

func NewDependencyHandler(_ Opts) handler.Handler {
	dependencyHandlerNew := &dependencyHandler{}

	return dependencyHandlerNew
}

func (s *dependencyHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	dependencyRecord, ok := obj.(*admiralApi.Dependency)
	if !ok {
		ctx.Log.Error("error casting Dependency object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	sourceIdentity := dependencyRecord.Spec.Source
	if len(sourceIdentity) == 0 {
		ctx.Log.Error("error Dependency has no source, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	ctx.Log.Str(logger.SourceAssetKey, sourceIdentity).Any(logger.DestinationAssetKey, dependencyRecord.Spec.Destinations).Debug("add to dependency cache")
	for _, dIdentity := range dependencyRecord.Spec.Destinations {
		cache.IdentityDependency.AddDependencyToIdentity(sourceIdentity, dIdentity)
		cache.IdentityDependency.AddDependentToIdentity(dIdentity, sourceIdentity)
	}
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *dependencyHandler) Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	dependencyRecord, ok := newObj.(*admiralApi.Dependency)
	oldDependencyRecord, oldRecordOk := oldObj.(*admiralApi.Dependency)
	if !ok {
		ctx.Log.Error("error casting Dependency object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	newDestinations := map[string]interface{}{}

	sourceIdentity := dependencyRecord.Spec.Source
	if len(sourceIdentity) == 0 {
		ctx.Log.Error("error Dependency has no source, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}
	for _, dIdentity := range dependencyRecord.Spec.Destinations {
		ctx.Log.Str(logger.SourceAssetKey, sourceIdentity).Str(logger.DestinationAssetKey, dIdentity).Debug("update to dependency cache")
		cache.IdentityDependency.AddDependencyToIdentity(sourceIdentity, dIdentity)
		cache.IdentityDependency.AddDependentToIdentity(dIdentity, sourceIdentity)
		newDestinations[dIdentity] = nil
	}

	if !options.IsCacheWarmedUp() {
		ctx.Log.Info("Cache not warmed up, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	if oldRecordOk {
		for _, dIdentity := range oldDependencyRecord.Spec.Destinations {
			delete(newDestinations, dIdentity)
		}
	}

	// If there are no new destinations, we can skip the traffic config handler
	if len(newDestinations) == 0 {
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// If there are any destinations that were added to the dependency, we need to trigger traffic config handler
	for dIdentity := range newDestinations {
		ctx.Log.Str(logger.WorkloadIdentifierKey, dIdentity).Info("New destination found, triggering handlers")
		tcHandler := traffic_config.NewTrafficConfigHandler()
		tcHandler.TriggerTrafficConfigHandlerForIdentity(ctx, dIdentity, statusChan)
	}

	// Closing the status on parent
	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *dependencyHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	_, ok := obj.(*admiralApi.Dependency)
	if !ok {
		ctx.Log.Error("error casting Dependency object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (s *dependencyHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}
