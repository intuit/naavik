package executor

import (
	"sync"
	"testing"
	"time"

	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockObject struct {
	metadata metav1.Object
}

func (mo mockObject) GetObjectMeta() metav1.Object {
	return mo.metadata
}

type asyncRunnerMock struct {
	sync.Mutex
	counterMap map[string]int
}

func (a *asyncRunnerMock) Add(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	objMeta, _ := meta.Accessor(obj)
	a.Lock()
	a.counterMap[objMeta.GetName()] += 1
	ctx.Log.Infof("Add called %d", a.counterMap[objMeta.GetName()])
	a.Unlock()
	time.Sleep(1 * time.Second)
}

func (a *asyncRunnerMock) Update(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	objMeta, _ := meta.Accessor(obj)
	a.Lock()
	a.counterMap[objMeta.GetName()] += 2
	ctx.Log.Infof("Add called %d", a.counterMap[objMeta.GetName()])
	a.Unlock()
	time.Sleep(1 * time.Second)
}

func (a *asyncRunnerMock) Delete(ctx context.Context, obj interface{}, statusChan chan controller.EventProcessStatus) {
	objMeta, _ := meta.Accessor(obj)
	a.Lock()
	a.counterMap[objMeta.GetName()] -= 1
	a.Unlock()
}

func (*asyncRunnerMock) GetKey(ctx context.Context, obj interface{}) string {
	objMeta, _ := meta.Accessor(obj)
	return objMeta.GetNamespace() + "/" + objMeta.GetName()
}

func mockStatusListener(obj mockObject, statusChan chan controller.EventProcessStatus) {
	for status := range statusChan {
		logger.Log.Infof("Status received obj=%s status=%s generation=%d", obj.metadata.GetName(), status.Status, obj.metadata.GetGeneration())
	}
}

func TestAsyncKeyExecutor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "async_key_executor_test")
}
