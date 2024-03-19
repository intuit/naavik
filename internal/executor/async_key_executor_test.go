package executor

import (
	"fmt"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("async_key_executor_test", func() {
	var async_key_executor Executor
	var async_runner_mock *asyncRunnerMock

	BeforeEach(func() {
		logger.Log.Info("Running before each")
		options.InitializeNaavikArgs(nil)
		async_runner_mock = &asyncRunnerMock{
			counterMap: map[string]int{},
		}
		async_key_executor = NewAsyncKeyExecutor("mock_executor", async_runner_mock)
	})

	When("concurrent events come for same object and the first event still in progress", func() {
		It("should execute event sequentially and ignore all intermediate generation and only process the last", func() {
			startTime := time.Now()
			var gen int64 = 0
			for i := 0; i < 100; i++ {
				ctx := context.NewContextWithLogger()
				ctx.Log.WithInt("goroutine", i)
				mockObject := mockObject{
					metadata: &metav1.PartialObjectMetadata{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test",
							Namespace:  "test",
							Generation: gen,
						},
					},
				}
				gen += 1
				statusCh := make(chan controller.EventProcessStatus)
				go mockStatusListener(mockObject, statusCh)
				async_key_executor.Add(ctx, mockObject, statusCh)
			}
			for time.Since(startTime) < 3*time.Second {
			}
			async_runner_mock.Lock()
			counterVal := async_runner_mock.counterMap["test"]
			async_runner_mock.Unlock()
			Expect(counterVal).To(Equal(2))
		})
	})

	When("concurrent events come for same object after the previous is finished processing", func() {
		It("should execute event sequentially", func() {
			startTime := time.Now()
			var gen int64 = 0
			for i := 0; i < 5; i++ {
				ctx := context.NewContextWithLogger()
				ctx.Log.WithInt("goroutine", i)
				mockObject := mockObject{
					metadata: &metav1.PartialObjectMetadata{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test",
							Namespace:  "test",
							Generation: gen,
						},
					},
				}
				gen += 1
				statusCh := make(chan controller.EventProcessStatus)
				go mockStatusListener(mockObject, statusCh)
				async_key_executor.Add(ctx, mockObject, statusCh)
				time.Sleep(1 * time.Second)
			}
			for time.Since(startTime) < 6*time.Second {
			}
			async_runner_mock.Lock()
			counterVal := async_runner_mock.counterMap["test"]
			async_runner_mock.Unlock()
			Expect(counterVal).To(Equal(5))
		})
	})

	When("concurrent events come and same object gets delete event in the end", func() {
		It("should execute delete handler and skip the intermediate updates", func() {
			startTime := time.Now()
			var gen int64 = 0
			for i := 0; i < 5; i++ {
				ctx := context.NewContextWithLogger()
				ctx.Log.WithInt("goroutine", i)
				mockObject := mockObject{
					metadata: &metav1.PartialObjectMetadata{
						ObjectMeta: metav1.ObjectMeta{
							Name:       "test",
							Namespace:  "test",
							Generation: gen,
						},
					},
				}
				statusCh := make(chan controller.EventProcessStatus)
				go mockStatusListener(mockObject, statusCh)
				// time.Sleep(100 * time.Millisecond)
				if i == 4 {
					async_key_executor.Delete(ctx, mockObject, statusCh)
				} else {
					gen += 1
					async_key_executor.Update(ctx, mockObject, statusCh)
				}
			}
			for time.Since(startTime) < 3*time.Second {
			}
			async_runner_mock.Lock()
			counterVal := async_runner_mock.counterMap["test"]
			async_runner_mock.Unlock()
			Expect(counterVal).To(Equal(1))
		})
	})

	When("concurrent events come for various objectss", func() {
		It("should execute different object concurrently and same object sequentially, skipping the itermediate generations", func() {
			startTime := time.Now()
			for i := 0; i < 10; i++ {
				var gen int64 = 0
				for j := 0; j < 10; j++ {
					ctx := context.NewContextWithLogger()
					ctx.Log.WithInt(fmt.Sprintf("goroutine-%d", i), j)
					mockObject := mockObject{
						metadata: &metav1.PartialObjectMetadata{
							ObjectMeta: metav1.ObjectMeta{
								Name:       fmt.Sprintf("%s-%d", "test", i),
								Namespace:  "test",
								Generation: gen,
							},
						},
					}
					gen += 1
					statusCh := make(chan controller.EventProcessStatus)
					go mockStatusListener(mockObject, statusCh)
					async_key_executor.Update(ctx, mockObject, statusCh)
				}
			}
			for time.Since(startTime) < 5*time.Second {
			}
			for i := 0; i < 10; i++ {
				async_runner_mock.Lock()
				counterVal := async_runner_mock.counterMap[fmt.Sprintf("%s-%d", "test", i)]
				async_runner_mock.Unlock()
				Expect(counterVal).To(Equal(4), fmt.Sprintf("%s-%d", "test", i))
			}
		})
	})
})
