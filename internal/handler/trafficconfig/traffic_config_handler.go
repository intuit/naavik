package trafficconfig

import (
	goctx "context"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/leasechecker"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
)

//nolint:revive
type TrafficConfigHandler interface {
	handler.Handler
	TriggerTrafficConfigHandlerForIdentity(ctx context.Context, identity string, statusChan chan controller.EventProcessStatus)
}

func NewTrafficConfigHandler() TrafficConfigHandler {
	return &DefaultTrafficConfigHandler{}
}

type Opts struct{}

type DefaultTrafficConfigHandler struct{}

func (tch *DefaultTrafficConfigHandler) Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	tc, ok := obj.(*admiralv1.TrafficConfig)
	if !ok {
		ctx.Log.Error("error casting TrafficConfig object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	tcUtil := utils.TrafficConfigUtil(tc)

	if tcUtil.IsDisabled() || tcUtil.IsIgnored() {
		ctx.Log.Bool(types.IsDisabledKey, tcUtil.IsDisabled()).Bool(options.GetResourceIgnoreLabel(), tcUtil.IsIgnored()).Info("Traffic config is disabled or ignored, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	cache.TrafficConfigCache.AddTrafficConfigToCache(tc)

	// Update the cache before checking for read only
	if leasechecker.IsReadOnly() {
		ctx.Log.Info("Updated cache, but in read only mode, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Don't process during cache warmup time as it might cause applying filters with invalid data
	if !options.IsCacheWarmedUp() {
		ctx.Log.Str(logger.ResourceIdentifierKey, tc.Name).Info("cache is not warmed up yet, skipping processing")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// handle rate limiting filter
	if options.IsFeatureEnabled(types.FeatureThrottleFilter) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureThrottleFilter.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Throttle filter processing started")
		HandleRateLimiter(newCtx, tc, types.Add)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Throttle filter processing completed")
	}

	if options.IsFeatureEnabled(types.FeatureVirtualService) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureVirtualService.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Virtual service processing started")
		HandleVirtualServiceForTrafficConfig(newCtx, tc, statusChan)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Virtual service processing completed")
	}

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (tch *DefaultTrafficConfigHandler) Updated(ctx context.Context, newObj interface{}, _ interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	tc, ok := newObj.(*admiralv1.TrafficConfig)
	if !ok {
		ctx.Log.Error("error casting TrafficConfig object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	tcUtil := utils.TrafficConfigUtil(tc)

	if tcUtil.IsDisabled() || tcUtil.IsIgnored() {
		ctx.Log.Bool(types.IsDisabledKey, tcUtil.IsDisabled()).Bool(options.GetResourceIgnoreLabel(), tcUtil.IsIgnored()).Info("Traffic config is disabled or ignored, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Update the cache before checking for read only
	if leasechecker.IsReadOnly() {
		ctx.Log.Info("Updated cache, but in read only mode, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Don't process during cache warmup time as it might cause applying filters with invalid data
	if !options.IsCacheWarmedUp() {
		ctx.Log.Str(logger.ResourceIdentifierKey, tc.Name).Info("cache is not warmed up yet, skipping processing")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// handle rate limiting filter
	if options.IsFeatureEnabled(types.FeatureThrottleFilter) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureThrottleFilter.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Throttle filter processing started")
		HandleRateLimiter(newCtx, tc, types.Update)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Throttle filter processing completed")
	}

	if options.IsFeatureEnabled(types.FeatureVirtualService) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureVirtualService.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Virtual service processing started")
		HandleVirtualServiceForTrafficConfig(newCtx, tc, statusChan)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Virtual service processing completed")
	}

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (tch *DefaultTrafficConfigHandler) Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	tc, ok := obj.(*admiralv1.TrafficConfig)
	if !ok {
		ctx.Log.Error("error casting TrafficConfig object, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	cache.TrafficConfigCache.DeleteTrafficConfigFromCache(tc)

	// Update the cache before checking for read only
	if leasechecker.IsReadOnly() {
		ctx.Log.Info("Updated cache, but in read only mode, skipping handling.")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// Don't process during cache warmup time as it might cause applying filters with invalid data
	if !options.IsCacheWarmedUp() {
		ctx.Log.Str(logger.ResourceIdentifierKey, tc.Name).Info("cache is not warmed up yet, skipping processing")
		return controller.NewEventProcessStatus().SkipClose(statusChan)
	}

	// handle rate limiting filter
	if options.IsFeatureEnabled(types.FeatureThrottleFilter) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureThrottleFilter.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Throttle filter processing started")
		HandleRateLimiter(newCtx, tc, types.Delete)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Throttle filter processing completed")
	}

	if options.IsFeatureEnabled(types.FeatureVirtualService) {
		startTime := time.Now()
		newCtx := context.NewContextWithLogger()
		newCtx.Log = ctx.Log.Str(logger.HandlerNameKey, types.FeatureVirtualService.String())
		newCtx.Log.Str("txId", utils.TrafficConfigUtil(tc).GetTransactionID()).Str("revision", utils.TrafficConfigUtil(tc).GetRevision()).Info("Virtual service processing started")
		HandleVirtualServiceForTrafficConfig(newCtx, tc, statusChan)
		newCtx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Virtual service processing completed")
	}

	return controller.NewEventProcessStatus().SkipClose(statusChan)
}

func (tch *DefaultTrafficConfigHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
}

func (tch *DefaultTrafficConfigHandler) TriggerTrafficConfigHandlerForIdentity(ctx context.Context, identity string, statusChan chan controller.EventProcessStatus) {
	ctx.Log.Str(logger.WorkloadIdentifierKey, identity).Info("Triggering traffic config handler for identity started")

	// Trigger traffic config handler for self
	tcEntry := cache.TrafficConfigCache.GetTrafficConfigEntry(identity)
	if tcEntry == nil {
		ctx.Log.Str(logger.WorkloadIdentifierKey, identity).Trace("No traffic config found for identity")
	} else {
		for env, tc := range tcEntry.EnvTrafficConfig {
			childCtx, childStatusChan := controller.NewEventProcessStatus().CreateChildEvent(ctx, tch.OnStatus, statusChan)
			childCtx.Log.Str(logger.WorkloadIdentifierKey, identity).Str(logger.NameKey, tc.Name).Str(logger.EnvKey, env).Info("Triggering traffic config handler for self")
			tch.Updated(childCtx, tc, nil, childStatusChan)
		}
	}

	// Trigger traffic config handler for all dependents traffic configs setting the source identity in the context
	// so that the traffic config will be handled only for the triggered source identity
	dependents := cache.IdentityDependency.GetDependentsForIdentity(identity)
	for _, dependent := range dependents {
		dependentTrafficConfigEntry := cache.TrafficConfigCache.GetTrafficConfigEntry(dependent)
		if dependentTrafficConfigEntry != nil {
			for env, tc := range dependentTrafficConfigEntry.EnvTrafficConfig {
				childCtx, childStatusChan := controller.NewEventProcessStatus().CreateChildEvent(ctx, tch.OnStatus, statusChan)
				// Add source identity to context so that the traffic config will be handled only for the triggered source identity
				childCtx.Context = goctx.WithValue(childCtx.Context, types.SourceIdentityKey, identity)
				childCtx.Log.Str(logger.WorkloadIdentifierKey, dependent).Str(logger.NameKey, tc.Name).Str(logger.EnvKey, env).Info("Triggering traffic config handler for dependent")
				tch.Updated(childCtx, tc, nil, childStatusChan)
			}
		} else {
			ctx.Log.Str(logger.WorkloadIdentifierKey, dependent).Trace("No traffic config found for dependent identity")
		}
	}
	ctx.Log.Str(logger.WorkloadIdentifierKey, identity).Info("Triggering traffic config handler for identity completed")
}
