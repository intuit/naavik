package controller

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
)

const (
	DefaultMaxRetryCount                  = 5
	DefaultEventStatusBufferedChannelSize = 5
)

type EventProcessStatus struct {
	MaxRetryCount int
	Status        EventStatus
	RetryCount    int
	Retry         bool
	RetryAfter    time.Duration
	Message       map[string]string
	Error         error

	ChildEventContext context.Context
	ChildEventChan    chan EventProcessStatus
	ChildOnStatus     func(ctx context.Context, status EventProcessStatus)
}

type (
	EventStatus            string
	TerminatingEventStatus EventStatus
)

const (
	EventCreateChild      EventStatus = "CreateChildEvent"
	EventSkip             EventStatus = "Skip"
	EventRetry            EventStatus = "Retry"
	EventMaxRetryReached  EventStatus = "MaxRetryReached"
	EventCompleted        EventStatus = "Completed"
	EventFailure          EventStatus = "Failure"
	EventProcessing       EventStatus = "Processing"
	EventPartialCompleted EventStatus = "PartialCompleted"
)

func (es EventStatus) String() string {
	return string(es)
}

// NewEventProcessStatus returns a new EventProcessStatus with default values
//
// # Default values:
//
//	Status: EventSuccess
//	MaxRetryCount: DefaultMaxRetryCount=5
func NewEventProcessStatus() EventProcessStatus {
	return EventProcessStatus{
		Status:        EventCompleted,
		Message:       make(map[string]string),
		MaxRetryCount: DefaultMaxRetryCount,
		Error:         fmt.Errorf(""),
	}
}

func NewEventProcessStatusWithStatus(eventStatus EventProcessStatus) EventProcessStatus {
	return eventStatus
}

func (eps EventProcessStatus) WithStatus(status EventStatus) EventProcessStatus {
	eps.Status = status
	return eps
}

func (eps EventProcessStatus) WithRetry() EventProcessStatus {
	eps.Retry = true
	eps.Status = EventRetry
	return eps
}

func (eps EventProcessStatus) WithMaxRetry(count int) EventProcessStatus {
	eps.MaxRetryCount = count
	return eps
}

func (eps EventProcessStatus) WithMessage(key string, value string) EventProcessStatus {
	eps.Message[key] = value
	return eps
}

func (eps EventProcessStatus) WithError(err error) EventProcessStatus {
	eps.Error = err
	return eps
}

// SkipClose must be called for a event only once to send skip status and close the channel.
func (eps EventProcessStatus) SkipClose(statusChan chan EventProcessStatus) EventStatus {
	if statusChan != nil {
		eps.Status = EventSkip
		statusChan <- eps
		close(statusChan)
		return eps.Status
	}
	return eps.Status
}

// Send sends the event status to the event status channel
// It does not close the event status channel and can be used to send multiple event statuses
// NOTE: Send can be called multiple times for a given event but SendClose can be called only once.
func (eps EventProcessStatus) Send(statusChan chan EventProcessStatus) EventStatus {
	if statusChan != nil {
		statusChan <- eps
		return eps.Status
	}
	return eps.Status
}

// SendClose must be called for every event only once to send the final event status and close the channel
// Calling SendClose more than once will result in a panic
// Use Send() if you do not want to close the channel and want to send multiple event statuses
// NOTE: Send() can be called multiple times for a given event but SendClose() can be called only once.
func (eps EventProcessStatus) SendClose(statusChan chan EventProcessStatus) EventStatus {
	if statusChan != nil {
		statusChan <- eps
		close(statusChan)
		return eps.Status
	}
	return eps.Status
}

// CreateChildEvent creates a child event with childEventId and returns the child event status channel.
// If statusChan is nil, it will return nil for childStatusChan.
// If childOnStatus is nil, it will use the parent event's OnStatus function.
func (eps EventProcessStatus) CreateChildEvent(ctx context.Context, childOnStatus func(ctx context.Context, status EventProcessStatus), statusChan chan EventProcessStatus) (childContext context.Context, childStatusChan chan EventProcessStatus) {
	eps.Status = EventCreateChild
	if statusChan != nil {
		eps.ChildEventContext = context.NewContextWithLogger()
		eps.ChildEventContext.Log = ctx.Log.Str(logger.ChildEventIDKey, uuid.New().String())
		eps.ChildOnStatus = childOnStatus
		eps.ChildEventChan = make(chan EventProcessStatus, DefaultEventStatusBufferedChannelSize)
		statusChan <- eps
	}
	return eps.ChildEventContext, eps.ChildEventChan
}
