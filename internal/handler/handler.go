package handler

import (
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types/context"
)

type Handler interface {
	Added(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	Updated(ctx context.Context, newObj interface{}, oldObj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	Deleted(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) controller.EventStatus
	OnStatus(ctx context.Context, eventStatus controller.EventProcessStatus)
}
