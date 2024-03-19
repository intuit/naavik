package executor

import (
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
)

type eventItem struct {
	key        string
	addTime    time.Time
	ctx        context.Context
	obj        interface{}
	statusChan chan controller.EventProcessStatus
	runFunc    func(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	eventType  types.EventType
}

type Executor interface {
	Add(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	Update(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	Delete(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
}

type AsyncKeyBasedRunner interface {
	Add(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	Update(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	Delete(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus)
	GetKey(ctx context.Context, obj interface{}) string
}
