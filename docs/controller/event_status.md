# Controller Event Status

To provide a seamless handling of events, the controller provides a status channel for each event that informer receives. The status channel is a buffered channel with a buffer limit of 5. It will call the Status method of the delegator when it receives a status on the channel of the event.

#### The status channel is useful for handling the following scenarios:
* When the event is processed, the controller can send a status to the status channel to indicate the status of the event.
* The processing of the event happens in a separate goroutine. The controller can send a status to the status channel to indicate the status of the event.
* OnStatus of the delegator is called when the status channel receives a status of the event. The delegator handler can then handle the status of the event to update the status of the resource or update the status to remote data store.

#### Things to note:
* The status channel is a buffered channel with a buffer limit of 5. If the buffer is full, the worker will block until the status is received.
* The status channel can only be closed once. Closing the already closed status channel will cause a panic.
* Ensure the status channel is closed in the handler. Not closing the status channel will cause a goroutines to be idle forever.

#### Every handler will implement the following interface:
```go
// internal/handlers/handler.go
type Handler interface {
	Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	OnStatus(ctx context.Context, eventStatus controller.EventProcessStatus)
}
```

Example: [deployment_handler.go](internal/handlers/deployment_handler.go)
Added, Updated, and Deleted methods are called when the informer receives the corresponding event with a `statusChan`. The handler can send a status to the status channel to indicate the status of the event. The OnStatus method is called when the status channel receives a status on the channel of the event. The handler can then handle the status of the event to update the status of the resource or update the status to remote data store.

Below are some examples of how to use the status channel to update the status of the resource:

```go
// internal/controller/event_status.go
const (
	EventSkip             EventStatus = "Skip"
	EventRetry            EventStatus = "Retry"
	EventMaxRetryReached  EventStatus = "MaxRetryReached"
	EventCompleted        EventStatus = "Completed"
	EventFailure          EventStatus = "Failure"
	EventProcessing       EventStatus = "Processing"
	EventPartialCompleted EventStatus = "PartialCompleted"
)
```

To contruct the status, use chaining methods:

- `WithStatus(EventStatus)` is used to set the status of the event. Default status is `EventCompleted`
- `WithMessage(key, value)` is used to set the message of the event. Default message is empty.
- `WithRetry()` is used to set the retry of the event. Default retry is false. This will add the event back to the informer queue and will be processed again.
- `WithMaxRetry(count int)` is used to set the max retry count of the event. Default retry count is 0.
- `WithError(err error)` is used to set the error of the event. Default error is nil.
- `Send()` is used to send a status to the status channel. It will call the OnStatus method of the handler. 
```go
controller.NewEventProcessStatus().Send(statusChan)
```
- `SkipClose()` and `SendClose()` are the terminal methods used to send last status of the event and close the status channel, stopping the status goroutine. `SkipClose` will send the last status of the event and will not call the OnStatus method of the handler. `SendClose` will send the last status of the event and will call the OnStatus method of the handler.
```go
controller.NewEventProcessStatus().SkipClose(statusChan)
controller.NewEventProcessStatus().SendClose(statusChan)
```

Example with chaining: 
```go
	controller.NewEventProcessStatus()WithStatus(controller.EventPartialCompleted).WithMessage("vs", "failed on cluster_a due to api server failure.").SkipClose(statusChan)
```