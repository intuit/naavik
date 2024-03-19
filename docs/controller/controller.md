# Controller

Controller is a delegator implementation of K8s controller. It is responsible for managing the lifecycle of the controller, including starting informers, starting workers, and handling status signals.

To create a controller for a specific resource, you need to implement the following deletegator interface :
```go
// internal/controller/controller.go
type Delegator interface {
	Added(context.Context, interface{}, chan EventProcessStatus)
	Updated(context.Context, interface{}, interface{}, chan EventProcessStatus)
	Deleted(context.Context, interface{}, chan EventProcessStatus)
	GetInformer() cache.SharedIndexInformer
	Status(context.Context, EventProcessStatus)
}
```
Example: [deployment_controller.go](internal/controller/deployment_controller.go)

* The GetInformer method returns the informer for the resource. 
* The Added, Updated, and Deleted methods are called when the informer receives the corresponding event. 
* The Status method is called when the status channel receives a status on the channel of the event.

* Controller takes care of starting the informer and workers, and handles the status channel. You can create a controller using `NewController()`. 
* Controller takes a optional `WorkerConcurrency` parameter, which is the number of workers to start. If not specified, the default value is configured by global cmd arg `options.GetWorkerConcurrency() (defaults to 1)`.
* Having multiple workers will start multiple worker goroutines to process events from informer in parallel. This is useful when processing events is a long running task, such as creating a resource in a remote system. 
* Controller also takes a optional `ResyncPeriod` parameter, which is the resync period of the informer. If not specified, the default value is configured by global cmd arg `options.GetResyncPeriod()`. Configuring it to 0 will disable resyncing.
```text
NOTE: Traffic Controller uses a resync period of 0, as it does not need to resync the state of the resource. It only needs to process the events from the informer.
```
* It also ensure every event from informer gets a unique eventId and is added to the logging context of the event. This is useful for debugging and tracing.
* Controller also spawns a goroutine to handle the status channel. The status channel has a buffer limit of 5. It will call the Status method of the delegator when it receives a status on the channel of the event. 