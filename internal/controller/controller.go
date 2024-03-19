package controller

import (
	contxt "context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/intuit/naavik/cmd/options"
	internalCache "github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"

	"k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Delegator interface contains the methods that are required.
type Delegator interface {
	Added(context.Context, interface{}, chan EventProcessStatus)
	Updated(context.Context, interface{}, interface{}, chan EventProcessStatus)
	Deleted(context.Context, interface{}, chan EventProcessStatus)
	GetInformer() cache.SharedIndexInformer
	Status(context.Context, EventProcessStatus)
}

type InformerCacheObj struct {
	key              string
	eventType        types.EventType
	obj              interface{}
	oldObj           interface{}
	retryCount       int
	addTime          time.Time
	processStart     time.Time
	statusChan       chan EventProcessStatus
	onStatusOverride func(context.Context, EventProcessStatus)
}

type Controller struct {
	Delegator
	name              string
	queue             workqueue.RateLimitingInterface
	informer          cache.SharedIndexInformer
	workerConcurrency int
}

type Opts struct {
	Name              string
	Context           context.Context
	Delegator         Delegator
	Informer          cache.SharedIndexInformer
	Queue             workqueue.RateLimitingInterface
	WorkerConcurrency int // Number of workers to run for the controller
}

// NewController creates a new controller with the given name, informer, and stop channel.
func NewController(opts Opts) Delegator {
	ctx := context.NewContextWithLogger()
	controller := &Controller{
		name:              opts.Name,
		Delegator:         opts.Delegator,
		queue:             workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		workerConcurrency: options.GetWorkerConcurrency(),
		informer:          opts.Delegator.GetInformer(),
	}
	// Override worker concurrency if provided else use global value from options.GetWorkerConcurrency()
	if opts.WorkerConcurrency > 0 {
		controller.workerConcurrency = opts.WorkerConcurrency
	}
	if opts.Informer != nil {
		controller.informer = opts.Informer
	}
	if opts.Context != (context.Context{}) {
		ctx = opts.Context
	}

	controller.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			eventStatus := make(chan EventProcessStatus, DefaultEventStatusBufferedChannelSize)
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				logger.NewLogger().WithStr(logger.ControllerNameKey, controller.name).With("obj", key).Info("Informer add event received")
				controller.queue.Add(&InformerCacheObj{key: key, eventType: types.Add, obj: obj, addTime: time.Now(), statusChan: eventStatus})
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			eventStatus := make(chan EventProcessStatus, DefaultEventStatusBufferedChannelSize)
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				logger.NewLogger().WithStr(logger.ControllerNameKey, controller.name).With("obj", key).Info("Informer update event received")
				controller.queue.Add(&InformerCacheObj{key: key, eventType: types.Update, obj: newObj, oldObj: oldObj, addTime: time.Now(), statusChan: eventStatus})
			}
		},
		DeleteFunc: func(obj interface{}) {
			eventStatus := make(chan EventProcessStatus, DefaultEventStatusBufferedChannelSize)
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				logger.NewLogger().WithStr(logger.ControllerNameKey, controller.name).With("obj", key).Info("Informer delete event received")
				controller.queue.Add(&InformerCacheObj{key: key, eventType: types.Delete, obj: obj, addTime: time.Now(), statusChan: eventStatus})
			}
		},
	})

	go controller.Run(ctx)

	return controller
}

// Run starts the controller's informer and the starts processing the items in the queue.
func (c *Controller) Run(ctx context.Context) {
	ctx.Log.WithStr(logger.ControllerNameKey, c.name).Info("Starting controller")

	stopCh := make(chan struct{})

	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	cancelContexts := []contxt.Context{}
	for i := 0; i < c.workerConcurrency; i++ {
		// Register the controller with the controller cache, used to stop the controller during termination
		cancelCtx, cancel := contxt.WithCancel(contxt.Background())
		cancelContexts = append(cancelContexts, cancelCtx)
		go c.runWorker(ctx, stopCh, cancel)
	}

	internalCache.ControllerCache.Register(c.name, internalCache.ControllerContext{WorkerCtx: cancelContexts, StopCh: stopCh})

	c.informer.Run(stopCh)

	ctx.Log.WithStr(logger.ControllerNameKey, c.name).Info("Informer stopped")
}

func (c *Controller) runWorker(ctx context.Context, stopCh chan struct{}, close contxt.CancelFunc) {
	// Close the controller context when the worker is done
	defer close()
	// Wait for the caches to be synced before starting workers
	ctx.Log.Info("Waiting for informer caches to sync")

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync for controller=%v", c.name))
		return
	}
	ctx.Log.Any("keys", c.informer.GetStore().ListKeys()).Info("Informer caches synced")

	for c.processNextItem(ctx) {
		// continue processing until the queue is shutdown
	}
	ctx.Log.Info("Process worker stopped")
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	item, quit := c.queue.Get()

	if quit {
		// Wait for the queue to drain,
		// we'll only wait for the active workers to complete the processing
		// and not process any new items
		ctx.Log.Info("Queue is shutdown, waiting for workers to drain")
		c.queue.ShutDownWithDrain()
		ctx.Log.Info("Queue is drained, stopping worker")
		return false
	}
	c.processItem(item.(*InformerCacheObj))

	return true
}

func (c *Controller) processItem(informerCacheObj *InformerCacheObj) {
	ctx := context.NewContextWithLogger()
	ctx.Log.WithStr(logger.EventIDKey, uuid.New().String()).
		Str(logger.ResourceIdentifierKey, informerCacheObj.key).
		Str(logger.EventType, informerCacheObj.eventType.String()).
		Str(logger.ControllerNameKey, c.name).
		Int(logger.QueueTimeMSKey, int(time.Since(informerCacheObj.addTime).Milliseconds())).
		Int(logger.QueueLengthKey, c.queue.Len()).
		Info("Processing started.")

	informerCacheObj.processStart = time.Now()

	// Initiate satatus callback handler
	c.handleStatus(ctx, informerCacheObj)

	if informerCacheObj.eventType == types.Delete {
		c.Deleted(ctx, informerCacheObj.obj, informerCacheObj.statusChan)
	} else if informerCacheObj.eventType == types.Update {
		c.Updated(ctx, informerCacheObj.obj, informerCacheObj.oldObj, informerCacheObj.statusChan)
	} else if informerCacheObj.eventType == types.Add {
		c.Added(ctx, informerCacheObj.obj, informerCacheObj.statusChan)
	}
}

func StopAllControllers() {
	var wg sync.WaitGroup
	internalCache.ControllerCache.Range(func(key, value interface{}) bool {
		wg.Add(1)
		go func(key, value interface{}) {
			defer wg.Done()
			startTime := time.Now()
			logger.Log.Str(logger.ControllerNameKey, key.(string)).Info("Stopping controller")
			ctlrContext := value.(internalCache.ControllerContext)
			close(ctlrContext.StopCh)
			for i, closeCtx := range ctlrContext.WorkerCtx {
				logger.Log.Str(logger.ControllerNameKey, key.(string)).Infof("Waiting for workers to finish... %d/%d", i+1, len(ctlrContext.WorkerCtx))
				<-closeCtx.Done()
			}
			logger.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Str(logger.ControllerNameKey, key.(string)).Info("Stopped controller")
		}(key, value)
		return true
	})
	wg.Wait()
}

// Status runs the status callback handler for the controller delegator and logs the time taken
// The call back is run in a go routine to avoid blocking the controller.
func (c *Controller) handleStatus(ctx context.Context, item *InformerCacheObj) {
	go func(ctx context.Context, item *InformerCacheObj) {
		defer c.queue.Done(item)
		startTime := time.Now()
		for eventStatus := range item.statusChan {
			ctx.Log.Str(logger.EventStatusKey, eventStatus.Status.String()).Info("OnStatus triggered.")
			if eventStatus.Status == EventCreateChild {
				newChildItem := &InformerCacheObj{
					key:          item.key,
					eventType:    item.eventType,
					obj:          item.obj,
					oldObj:       item.oldObj,
					retryCount:   item.retryCount,
					addTime:      time.Now(),
					processStart: item.processStart,
					// Use the new child event status channel and onStatus callback
					statusChan:       eventStatus.ChildEventChan,
					onStatusOverride: eventStatus.ChildOnStatus,
				}
				c.handleStatus(eventStatus.ChildEventContext, newChildItem)
			} else if eventStatus.Status == EventSkip {
				ctx.Log.Trace("Skipping OnStatus callback")
			} else if (eventStatus.Status == EventCompleted || eventStatus.Status == EventFailure || eventStatus.Status == EventPartialCompleted) && !eventStatus.Retry {
				c.triggerOnStatusCallback(ctx, eventStatus, item)
			} else if eventStatus.Retry && item.retryCount < eventStatus.MaxRetryCount {
				ctx.Log.Errorf("event will be retried. %d/%d", item.retryCount, eventStatus.MaxRetryCount)
				eventStatus.RetryCount = item.retryCount
				c.triggerOnStatusCallback(ctx, eventStatus, item)
				// Increment the retry count
				item.retryCount++
				// Recreate the status channel to avoid sending to a closed channel
				item.statusChan = make(chan EventProcessStatus, DefaultEventStatusBufferedChannelSize)
				if eventStatus.RetryAfter > 0 {
					c.queue.AddAfter(item, eventStatus.RetryAfter)
				} else {
					c.queue.Add(item)
				}
			} else if eventStatus.Retry && item.retryCount >= eventStatus.MaxRetryCount {
				eventStatus.RetryCount = item.retryCount
				eventStatus.Status = EventMaxRetryReached
				c.triggerOnStatusCallback(ctx, eventStatus, item)
				ctx.Log.Errorf("error processing item max retry reached %d/%d, giving up", item.retryCount, eventStatus.MaxRetryCount)
			} else {
				c.triggerOnStatusCallback(ctx, eventStatus, item)
			}
			ctx.Log.Int(logger.TimeTakenMSKey, int(time.Since(startTime).Milliseconds())).Str(logger.EventStatusKey, eventStatus.Status.String()).Info("OnStatus completed.")
		}
		ctx.Log.Str(logger.ControllerNameKey, c.name).Str(logger.ResourceIdentifierKey, item.key).Any(logger.TimeTakenMSKey, time.Since(item.processStart).Milliseconds()).Info("Processing completed.")
	}(ctx, item)
}

func (c *Controller) triggerOnStatusCallback(ctx context.Context, eventStatus EventProcessStatus, informerCacheObj *InformerCacheObj) {
	// If the onStatus callback is overridden, use the overridden callback else use the delegator callback
	if informerCacheObj.onStatusOverride != nil {
		informerCacheObj.onStatusOverride(ctx, eventStatus)
	} else {
		c.Status(ctx, eventStatus)
	}
}
