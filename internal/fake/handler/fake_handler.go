package handler

import (
	"sync/atomic"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
)

// Mock Handler.
type FakeHandler struct {
	handler.Handler
	clusterID     string
	sleep         time.Duration
	eventStatus   *controller.EventProcessStatus
	childEvent    bool
	childOnStatus func(ctx context.Context, status controller.EventProcessStatus)

	AddCalled      atomic.Int64
	UpdateCalled   atomic.Int64
	DeleteCalled   atomic.Int64
	OnStatusCalled atomic.Int64
}

func NewFakeNoOpHandler(clusterID string, sleep time.Duration) *FakeHandler {
	newFakeHandler := &FakeHandler{
		clusterID: clusterID,
		sleep:     sleep,
	}
	return newFakeHandler
}

func NewFakeNoOpHandlerWithStatus(clusterID string, eventStatus *controller.EventProcessStatus) *FakeHandler {
	newFakeHandler := &FakeHandler{
		clusterID:   clusterID,
		eventStatus: eventStatus,
	}
	return newFakeHandler
}

func NewFakeNoOpHandlerWithChildEvent(clusterID string, childOnStatus func(ctx context.Context, status controller.EventProcessStatus)) *FakeHandler {
	newFakeHandler := &FakeHandler{
		clusterID:     clusterID,
		childEvent:    true,
		childOnStatus: childOnStatus,
	}
	return newFakeHandler
}

func (d *FakeHandler) Added(ctx context.Context, _ interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	d.AddCalled.Add(1)
	if d.sleep > 0 {
		time.Sleep(d.sleep)
	}
	if d.eventStatus != nil {
		return controller.NewEventProcessStatusWithStatus(*d.eventStatus).SendClose(statusChan)
	}
	if d.childEvent {
		_, childStatusChan := controller.NewEventProcessStatus().CreateChildEvent(ctx, d.childOnStatus, statusChan)
		controller.NewEventProcessStatus().SendClose(childStatusChan)
	}
	return controller.NewEventProcessStatus().SendClose(statusChan)
}

func (d *FakeHandler) Updated(ctx context.Context, _ interface{}, _ interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	d.UpdateCalled.Add(1)
	if d.sleep > 0 {
		time.Sleep(d.sleep)
	}
	if d.eventStatus != nil {
		return controller.NewEventProcessStatusWithStatus(*d.eventStatus).SendClose(statusChan)
	}
	if d.childEvent {
		_, childStatusChan := controller.NewEventProcessStatus().CreateChildEvent(ctx, d.childOnStatus, statusChan)
		controller.NewEventProcessStatus().SendClose(childStatusChan)
	}
	return controller.NewEventProcessStatus().SendClose(statusChan)
}

func (d *FakeHandler) Deleted(ctx context.Context, _ interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus {
	d.DeleteCalled.Add(1)
	if d.sleep > 0 {
		time.Sleep(d.sleep)
	}
	if d.eventStatus != nil {
		return controller.NewEventProcessStatusWithStatus(*d.eventStatus).SendClose(statusChan)
	}
	if d.childEvent {
		_, childStatusChan := controller.NewEventProcessStatus().CreateChildEvent(ctx, d.childOnStatus, statusChan)
		controller.NewEventProcessStatus().SendClose(childStatusChan)
	}
	return controller.NewEventProcessStatus().SendClose(statusChan)
}

func (d *FakeHandler) OnStatus(_ context.Context, _ controller.EventProcessStatus) {
	d.OnStatusCalled.Add(1)
}
