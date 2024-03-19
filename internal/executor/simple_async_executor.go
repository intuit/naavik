//go:build exclude

// This is still in alpha stage and not validated for use in production
package executor

import (
	"fmt"
	"sync"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
)

var eventsInProgress = map[string]*workerData{} // sync key to worker data map

var workerMap = map[string]*workerData{} // workerID to worker data map

type workerData struct {
	sync.Mutex
	workerID string
	queue    chan eventItem
}

type simpleAsyncExecutor struct {
	name   string
	runner AsyncKeyBasedRunner
}

// This is still in alpha stage and not validated for use in production
func StartSimpleAsyncExecutor(name string, noOfWorkers int, runner AsyncKeyBasedRunner) Executor {
	sae := &simpleAsyncExecutor{
		name:   name,
		runner: runner,
	}
	// Create workers
	for i := 1; i < noOfWorkers+1; i++ {
		workerID := fmt.Sprintf("worker-%s-%d", name, i)
		worker := &workerData{
			workerID: workerID,
			queue:    make(chan eventItem, 50),
		}
		workerMap[workerID] = worker
		logger.Log.Str(logger.NameKey, workerID).Infof("Starting async worker for %s", name)
		sae.startWorker(worker)
	}

	return sae
}

func (ae *simpleAsyncExecutor) Add(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Add)
}

func (ae *simpleAsyncExecutor) Update(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Update)
}

func (ae *simpleAsyncExecutor) Delete(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	ae.execute(ctx, obj, statusChan, ae.runner.Delete)
}

// The execute checks if the obj key is process by any of the worker go routine
// If yes then it will push the obj to the worker queue
// If no then it will find a worker with lease items in queue and push the obj to that worker queue
func (ae *simpleAsyncExecutor) execute(
	ctx context.Context,
	obj interface{},
	statusChan chan controller.EventProcessStatus,
	runFunc func(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus),
) {
	// Get the obj key for the obj
	objKey := ae.runner.GetKey(ctx, obj)

	// Get the worker for the key if any worker is already processing the key
	worker, err := ae.getWorker(ctx, objKey)
	if err != nil {
		ctx.Log.Errorf("unable to add event to worker, adding it back to informer queue. %s", err)
		controller.NewEventProcessStatus().WithRetry().SendClose(statusChan)
		return
	}

	ctx.Log.Str(logger.NameKey, worker.workerID).Int(logger.QueueLengthKey, len(worker.queue)).Trace("Assiging event to async worker")
	worker.queue <- eventItem{
		key:        objKey,
		addTime:    time.Now(),
		ctx:        ctx,
		obj:        obj,
		statusChan: statusChan,
		runFunc:    runFunc,
	}
	ctx.Log.Str(logger.NameKey, worker.workerID).Int(logger.QueueLengthKey, len(worker.queue)).Info("Assigned event to async worker")
}

func (ae *simpleAsyncExecutor) getWorker(ctx context.Context, key string) (*workerData, error) {
	// Check if the sync key is already being processed by any of the worker
	worker, ok := eventsInProgress[key]
	if ok {
		if len(worker.queue) >= 3 {
			return nil, fmt.Errorf("worker queue length reached max limit %d", len(worker.queue))
		}
		ctx.Log.Str(logger.NameKey, worker.workerID).Str(logger.ResourceIdentifierKey, key).Info("Worker already processing event, adding to same worker queue")
	} else {
		// If not then find a worker with lease items in queue
		worker = ae.findWorkerWithLeasetQueuedItems()
	}
	return worker, nil
}

func (ae *simpleAsyncExecutor) findWorkerWithLeasetQueuedItems() *workerData {
	var minQueueWorker *workerData
	// Find a worker with lease items in queue
	for _, worker := range workerMap {
		if minQueueWorker == nil {
			minQueueWorker = worker
		} else if len(worker.queue) < len(minQueueWorker.queue) {
			minQueueWorker = worker
		}
	}
	return minQueueWorker
}

// The startWorker starts a go routine for the worker
func (ae *simpleAsyncExecutor) startWorker(worker *workerData) {
	go func() {
		for {
			worker.Lock()
			select {
			case event := <-worker.queue:
				worker.Unlock()
				startTime := time.Now()
				event.ctx.Log.Str(logger.NameKey, worker.workerID).
					Int(logger.QueueLengthKey, len(worker.queue)).
					Int(logger.QueueTimeMSKey, int(time.Since(event.addTime))).
					Info("Processing event from async worker")

				// Add the key to eventsInProgress worker map
				eventsInProgress[event.key] = worker

				// Execute the run function
				event.runFunc(event.ctx, event.obj, event.statusChan)

				// Remove the sync key from eventsInProgress map
				delete(eventsInProgress, event.key)

				event.ctx.Log.Str(logger.NameKey, worker.workerID).
					Int(logger.TimeTakenMSKey, int(time.Since(startTime))).
					Info("Processed event from async worker")
			}
		}
	}()
}
