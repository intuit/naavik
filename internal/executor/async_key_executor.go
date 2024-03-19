package executor

import (
	goctx "context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
)

type asyncExecutor struct {
	name                     string
	runner                   AsyncKeyBasedRunner
	runnerKeyMap             sync.Map
	totalAsyncRunsInProgress atomic.Int32
}

type runnerData struct {
	sync.Mutex

	queueSlot     sync.Mutex
	toBeProcessed *eventItem
	lastProcessed *eventItem
}

// NewAsyncKeyExecutor creates a new async executor.
// The async executor executes the async function in a go routine and maintains a map of sync key to runner data
// Executor executes the run function sequentially for a given sync key
// If the executor receives a new event for a sync key while the previous event is still being processed,
// the new event will wait for lock to be released from prev event and then execute the run function.
func NewAsyncKeyExecutor(name string, runner AsyncKeyBasedRunner) Executor {
	return &asyncExecutor{
		name:         name,
		runner:       runner,
		runnerKeyMap: sync.Map{}, // sync key to runner data map
	}
}

// Add executes the runner Add function sequentially for a given sync key.
func (ae *asyncExecutor) Add(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Add)
}

// Update executes the runner Update function sequentially for a given sync key.
func (ae *asyncExecutor) Update(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Update)
}

// Delete executes the runner Delete function sequentially for a given sync key.
func (ae *asyncExecutor) Delete(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Delete)
}

// Donot touch the mutex locks unless you are sure what you are doing
// TODO: cognitive score is 31, improve to make it below 20
//
//nolint:gocognit
func (ae *asyncExecutor) execute(
	ctx context.Context,
	obj interface{},
	statusChan chan controller.EventProcessStatus,
	runFunc func(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus),
) {
	ctx.Log.Str(logger.NameKey, ae.name).Info("Initiating async key executor in a go routine")
	ae.waitForMaxAsyncRunsInProgress(ctx)
	ae.totalAsyncRunsInProgress.Add(1)

	// Execute the async function in a go routine
	go func(ae *asyncExecutor, ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
		syncKey := ae.runner.GetKey(ctx, obj)
		if len(syncKey) == 0 {
			ctx.Log.Error("sync key is empty, skipping execution")
			controller.NewEventProcessStatus().SkipClose(statusChan)
			return
		}
		defer ae.totalAsyncRunsInProgress.Add(-1)
		startTime := time.Now()
		item, _ := ae.runnerKeyMap.LoadOrStore(syncKey, &runnerData{})
		runnerData := item.(*runnerData)
		// Check if the event is already in processing for the given sync key
		if ae.isEventAlreadyInProcessingForKey(runnerData) {
			// runnerData.queueSlot Lock 1
			runnerData.queueSlot.Lock()
			// Check if there is any item in the queue slot, if not add the item to the queue slot and wait for the previous event to finish
			toBeProcessed := runnerData.toBeProcessed
			if toBeProcessed == nil {
				ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Info("No item in queue, adding item to queue")
				ae.addItemToQueueSlot(ctx, runnerData, syncKey, obj, statusChan, runFunc)

				// runnerData.queueSlot Unlock 1
				runnerData.queueSlot.Unlock()

				// Wait for the previous event to finish and then resume processing by picking the item from the queue slot
				// runnerData Lock 2
				runnerData.Lock()
				// runnerData Unlock 2
				defer runnerData.Unlock()

				// runnerData.queueSlot Lock 2
				runnerData.queueSlot.Lock()
				// Check if last processed object generation is higher than the present one
				ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Trace("Validating the generation of previous processed object")
				if runnerData.lastProcessed != nil && runnerData.toBeProcessed != nil {
					lastProcessed, errO := meta.Accessor(runnerData.lastProcessed.obj)
					newObjMetadata, errN := meta.Accessor(runnerData.toBeProcessed.obj)
					if errO != nil || errN != nil {
						// Ideally we should never hit this case
						ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Errorf("error getting object metadata while validating lastProcessed object, skipping %v", errO)
						controller.NewEventProcessStatus().SkipClose(runnerData.toBeProcessed.statusChan)
						runnerData.toBeProcessed = nil
						// runnerData.queueSlot Unlock 2
						runnerData.queueSlot.Unlock()
						return
					} else if lastProcessed.GetGeneration() > newObjMetadata.GetGeneration() {
						ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Warnf("Last processed object generation=%d is higher than present queued generation=%d, skipping", lastProcessed.GetGeneration(), newObjMetadata.GetGeneration())
						controller.NewEventProcessStatus().SkipClose(runnerData.toBeProcessed.statusChan)
						runnerData.toBeProcessed = nil
						// runnerData.queueSlot Unlock 2
						runnerData.queueSlot.Unlock()
						return
					}
				}

				// Get the item from the queue slot
				if runnerData.toBeProcessed != nil {
					ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Info("Picking item from queue")
					ctx = runnerData.toBeProcessed.ctx
					obj = runnerData.toBeProcessed.obj
					statusChan = runnerData.toBeProcessed.statusChan
					runFunc = runnerData.toBeProcessed.runFunc
					runnerData.toBeProcessed = nil
				}
				// runnerData.queueSlot Unlock 2
				runnerData.queueSlot.Unlock()
			} else {
				// If the queue slot is not empty, check if the new event has higher generation than the event in the queue slot
				// If the new event has higher generation, skip processing the event in the queue slot and process the new event
				// If the new event has lower generation, skip processing the new event and process the event in the queue slot
				toBeProcessedObjMetadata, errO := meta.Accessor(toBeProcessed.obj)
				newObjMetadata, errN := meta.Accessor(obj)
				if errO != nil || errN != nil {
					// Ideally we should never hit this case
					ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Errorf("error getting object metadata, skipping %v", errO)
					controller.NewEventProcessStatus().SkipClose(statusChan)
				} else if toBeProcessedObjMetadata.GetGeneration() <= newObjMetadata.GetGeneration() {
					toBeProcessed.ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Warnf("Higher generation object received %d during in queue, skipping processing of the old generation %d", newObjMetadata.GetGeneration(), toBeProcessedObjMetadata.GetGeneration())
					oldChan := toBeProcessed.statusChan
					controller.NewEventProcessStatus().SkipClose(oldChan)
					ctx.Log.Str(logger.NameKey, ae.name).Infof("Adding the higher generation %d object to the queue", newObjMetadata.GetGeneration())
					ae.addItemToQueueSlot(ctx, runnerData, syncKey, obj, statusChan, runFunc)
				} else {
					ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Warnf("Lower generation=%d object received, when item with generation=%d in queue, skipping processing old lower gen.", newObjMetadata.GetGeneration(), toBeProcessedObjMetadata.GetGeneration())
					controller.NewEventProcessStatus().SkipClose(statusChan)
				}
				ctx.Log.Str(logger.NameKey, ae.name).Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Info("Async executor finished")

				// runnerData.queueSlot Unlock 1
				runnerData.queueSlot.Unlock()
				// Exit from the go routine
				return
			}
		} else {
			// runnerData Unlock 1
			defer runnerData.Unlock()
		}

		ctx.Log.Str(logger.NameKey, ae.name).Str("syncKey", syncKey).Info("Executing async executor")
		queueTime := time.Since(startTime)

		// Execute the async function
		runFunc(ctx, obj, statusChan)

		// Update last processed object
		runnerData.lastProcessed = &eventItem{
			ctx:        ctx,
			obj:        obj,
			statusChan: statusChan,
			runFunc:    runFunc,
		}

		ctx.Log.Str(logger.NameKey, ae.name).
			Int(logger.WaitTimeMSKey, int(queueTime.Milliseconds())).
			Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).
			Infof("Async executor finished")
	}(ae, ctx, obj, statusChan)
}

func (ae *asyncExecutor) isEventAlreadyInProcessingForKey(runnerData *runnerData) bool {
	// runnerData Lock 1
	return !runnerData.TryLock()
}

func (ae *asyncExecutor) addItemToQueueSlot(
	ctx context.Context,
	runnerData *runnerData,
	syncKey string,
	obj interface{},
	statusChan chan controller.EventProcessStatus,
	runFunc func(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus),
) {
	runnerData.toBeProcessed = &eventItem{
		key:        syncKey,
		addTime:    time.Now(),
		ctx:        ctx,
		obj:        obj,
		statusChan: statusChan,
		runFunc:    runFunc,
	}
}

// Wait till the number of async runs in progress is less than maxAsyncRuns.
func (ae *asyncExecutor) waitForMaxAsyncRunsInProgress(ctx context.Context) {
	wait.PollUntilContextCancel(goctx.WithValue(ctx.Context, "logger", ctx.Log), time.Second, true,
		func(fctx goctx.Context) (done bool, err error) {
			done = ae.totalAsyncRunsInProgress.Load() < options.GetAsyncExecutorMaxGoRoutines()
			if !done {
				log, ok := fctx.Value("logger").(logger.Logger)
				if ok {
					log.Int("totalAsyncRunsInProgress", int(ae.totalAsyncRunsInProgress.Load())).Warn("Too many go routines in progress, waiting for some to finish")
				}
			}
			return done, nil
		})
}
