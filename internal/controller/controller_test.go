package controller_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	fake_builder "github.com/intuit/naavik/internal/fake/builder/resource"
	fake_controller "github.com/intuit/naavik/internal/fake/controller"
	fake_handler "github.com/intuit/naavik/internal/fake/handler"
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	"github.com/intuit/naavik/internal/types/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ = Describe("Test controller", Label("deployments_cache_test"), func() {
	When("New controller is created", func() {
		var config *rest.Config
		var fakeClusterName string

		BeforeEach(func() {
			options.InitializeNaavikArgs(nil)
			fakeClusterName = "fake_cluster"
			config, _ = fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(fakeClusterName)
		})

		AfterEach(func() {
			fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
			controller.StopAllControllers()
			cache.ControllerCache.Reset()
		})

		It("should start new controller and informer. Handler should receive events on resource change", func() {
			controllerName := fmt.Sprintf("mock-controller/%s", config.Host)
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			mockNamesapce := "fake_namespace"
			listOpts := metav1.ListOptions{}

			handler := fake_handler.NewFakeNoOpHandler(config.ServerName, 0)
			fakeController := &fake_controller.FakeController{
				Clientset: client,
				Namespace: mockNamesapce,
				ListOpts:  listOpts,
				Handler:   handler,
			}
			informer := fakeController.GetInformer()
			controller.NewController(controller.Opts{
				Name:      controllerName,
				Delegator: fakeController,
				Informer:  informer,
			})
			Expect(informer).ToNot(BeNil())
			Eventually(informer.HasSynced, 5*time.Second).Should(BeTrue())

			// Creates a fake deployment on the fake cluster and verifies that the add handler is called
			dep1 := fake_builder.BuildFakeDeployment("fake_deployment-1", "intuit.app", "app", "env", mockNamesapce)
			dep2 := fake_builder.BuildFakeDeployment("fake_deployment-2", "ntuit.app", "app", "env", mockNamesapce)
			client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep1, metav1.CreateOptions{})
			client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep2, metav1.CreateOptions{})
			Eventually(handler.AddCalled.Load, 5*time.Second).Should(Equal(int64(2)), "Add handler should be called")
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(2)), "Status handler should be called")

			// Updates the fake deployment on the fake cluster and verifies that the update handler is called
			dep1.Annotations["new_annotation"] = "update_annotation"
			dep2.Annotations["new_annotation"] = "update_annotation"
			client.AppsV1().Deployments(mockNamesapce).Update(context.Background(), dep1, metav1.UpdateOptions{})
			client.AppsV1().Deployments(mockNamesapce).Update(context.Background(), dep2, metav1.UpdateOptions{})
			Eventually(handler.UpdateCalled.Load, 5*time.Second).Should(Equal(int64(2)), "Update handler should be called")
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(4)), "Status handler should be called")

			client.AppsV1().Deployments(mockNamesapce).Delete(context.Background(), dep1.Name, metav1.DeleteOptions{})
			client.AppsV1().Deployments(mockNamesapce).Delete(context.Background(), dep2.Name, metav1.DeleteOptions{})
			Eventually(handler.DeleteCalled.Load, 5*time.Second).Should(Equal(int64(2)), "Delete handler should be called")
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(6)), "Status handler should be called")
		})

		It("should start new controller and should get registered in controller cache", func() {
			// Verify that the no controllers are not registered in the cache
			Expect(len(cache.ControllerCache.List())).Should(Equal(0))

			controllerName := fmt.Sprintf("mock-controller/%s", config.Host)
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			mockNamesapce := "fake_namespace"
			listOpts := metav1.ListOptions{}

			handler := fake_handler.NewFakeNoOpHandler(config.ServerName, 0)
			fakeController := &fake_controller.FakeController{
				Clientset: client,
				Namespace: mockNamesapce,
				ListOpts:  listOpts,
				Handler:   handler,
			}
			informer := fakeController.GetInformer()
			controller.NewController(controller.Opts{
				Name:      controllerName,
				Delegator: fakeController,
				Informer:  informer,
			})
			Expect(informer).ToNot(BeNil())
			Eventually(informer.HasSynced, 5*time.Second).Should(BeTrue())
			Eventually(len(cache.ControllerCache.List())).Should(Equal(1))
		})
	})

	When("Handler creates child event with its context", func() {
		var config *rest.Config
		var fakeClusterName string

		BeforeEach(func() {
			options.InitializeNaavikArgs(nil)
			fakeClusterName = "fake_cluster"
			config, _ = fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(fakeClusterName)
		})

		AfterEach(func() {
			fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
			controller.StopAllControllers()
			cache.ControllerCache.Reset()
		})

		It("should trigger the child status handler when event pushed to child event channel", func() {
			controllerName := fmt.Sprintf("mock-controller/%s", config.Host)
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			mockNamesapce := "fake_namespace"
			listOpts := metav1.ListOptions{}

			// Create child event
			called := atomic.Bool{}

			childOnStatusFunc := func(ctx context.Context, status controller.EventProcessStatus) {
				ctx.Log.Info("Child event triggered")
				called.Store(true)
			}

			handler := fake_handler.NewFakeNoOpHandlerWithChildEvent(config.ServerName, childOnStatusFunc)
			fakeController := &fake_controller.FakeController{
				Clientset: client,
				Namespace: mockNamesapce,
				ListOpts:  listOpts,
				Handler:   handler,
			}
			informer := fakeController.GetInformer()
			controller.NewController(controller.Opts{
				Name:      controllerName,
				Delegator: fakeController,
				Informer:  informer,
			})
			Expect(informer).ToNot(BeNil())
			Eventually(informer.HasSynced, 5*time.Second).Should(BeTrue())

			// Creates a fake deployment on the fake cluster and verifies that the add handler is called
			dep1 := fake_builder.BuildFakeDeployment("fake_deployment-1", "intuit.app", "app", "env", mockNamesapce)
			client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep1, metav1.CreateOptions{})
			Eventually(called.Load, 5*time.Second).Should(BeTrue(), "Child status handler should be triggered")
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(1)), "Parent Status handler should be triggered")
		})
	})

	When("handler sends an event status to retry the obj must be retried", func() {
		var config *rest.Config
		var fakeClusterName string

		BeforeEach(func() {
			options.InitializeNaavikArgs(nil)
			fakeClusterName = "fake_cluster"
			config, _ = fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(fakeClusterName)
		})

		AfterEach(func() {
			fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
			controller.StopAllControllers()
			cache.ControllerCache.Reset()
		})

		It("should trigger handler N times based on retries configured and status also should be triggered N times", func() {
			controllerName := fmt.Sprintf("mock-controller/%s", config.Host)
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			mockNamesapce := "fake_namespace"
			listOpts := metav1.ListOptions{}
			eventStatus := controller.NewEventProcessStatus().WithRetry()
			handler := fake_handler.NewFakeNoOpHandlerWithStatus(config.ServerName, &eventStatus)
			fakeController := &fake_controller.FakeController{
				Clientset: client,
				Namespace: mockNamesapce,
				ListOpts:  listOpts,
				Handler:   handler,
			}
			informer := fakeController.GetInformer()
			controller.NewController(controller.Opts{
				Name:      controllerName,
				Delegator: fakeController,
				Informer:  informer,
			})
			Expect(informer).ToNot(BeNil())
			Eventually(informer.HasSynced, 5*time.Second).Should(BeTrue())

			// Creates a fake deployment on the fake cluster and verifies that the add handler is called
			dep1 := fake_builder.BuildFakeDeployment("fake_deployment-1", "intuit.app", "app", "env", mockNamesapce)
			client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep1, metav1.CreateOptions{})
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(6)), "Parent Status handler should be triggered")
		})
	})

	When("handler sends an event with status skip", func() {
		var config *rest.Config
		var fakeClusterName string

		BeforeEach(func() {
			options.InitializeNaavikArgs(nil)
			fakeClusterName = "fake_cluster"
			config, _ = fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(fakeClusterName)
		})

		AfterEach(func() {
			fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
			controller.StopAllControllers()
			cache.ControllerCache.Reset()
		})

		It("should not trigger status handler and skip", func() {
			controllerName := fmt.Sprintf("mock-controller/%s", config.Host)
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			mockNamesapce := "fake_namespace"
			listOpts := metav1.ListOptions{}
			eventStatus := controller.NewEventProcessStatus()
			eventStatus.Status = controller.EventSkip
			handler := fake_handler.NewFakeNoOpHandlerWithStatus(config.ServerName, &eventStatus)
			fakeController := &fake_controller.FakeController{
				Clientset: client,
				Namespace: mockNamesapce,
				ListOpts:  listOpts,
				Handler:   handler,
			}
			informer := fakeController.GetInformer()
			controller.NewController(controller.Opts{
				Name:      controllerName,
				Delegator: fakeController,
				Informer:  informer,
			})
			Expect(informer).ToNot(BeNil())
			Eventually(informer.HasSynced, 5*time.Second).Should(BeTrue())

			// Creates a fake deployment on the fake cluster and verifies that the add handler is called
			dep1 := fake_builder.BuildFakeDeployment("fake_deployment-1", "intuit.app", "app", "env", mockNamesapce)
			client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep1, metav1.CreateOptions{})
			Eventually(handler.OnStatusCalled.Load, 5*time.Second).Should(Equal(int64(0)), "Parent Status handler should not be triggered")
		})
	})

	When("When graceful termination is triggered", func() {
		BeforeEach(func() {
			options.InitializeNaavikArgs(nil)
		})
		AfterEach(func() {
			fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
			cache.ControllerCache.Reset()
		})

		It("should terminate all the controllers gracefully", func() {
			mockNamesapce := "fake_namespace"
			numberOfCluster := 10
			clusterHandlers := map[string]*fake_handler.FakeHandler{}
			clusterConfigs := map[string]*rest.Config{}
			// Verify that the no controllers are not registered in the cache
			Expect(len(cache.ControllerCache.List())).Should(Equal(0))
			controllerLogMessages := map[string][]string{}
			var logMessageLock sync.Mutex
			for i := 0; i < numberOfCluster; i++ {
				fakeClusterName := fmt.Sprintf("fake_cluster-%d", i)
				config, _ := fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(fakeClusterName)
				clusterConfigs[fakeClusterName] = config
				clusterHandlers[fakeClusterName] = fake_handler.NewFakeNoOpHandler(config.ServerName, 1*time.Second)
				ctx := context.NewContextWithLogger()
				logMessageLock.Lock()
				controllerLogMessages[fakeClusterName] = []string{}
				logMessageLock.Unlock()
				ctx.Log.Hook(func(_ string, msg string) {
					logMessageLock.Lock()
					defer logMessageLock.Unlock()
					controllerLogMessages[fakeClusterName] = append(controllerLogMessages[fakeClusterName], msg)
				})
				fake_controller.NewFakeController(ctx, fmt.Sprintf("fake_controller-%d", i),
					config, fake_k8s_utils.NewFakeConfigLoader(), mockNamesapce,
					metav1.ListOptions{}, 0,
					clusterHandlers[fakeClusterName])
			}
			// Give some time for informers to sync
			time.Sleep(2 * time.Second)
			// Verify that the controllers are registered in the cache
			Eventually(cache.ControllerCache.List).Should(HaveLen(numberOfCluster))

			// Push continuous events to one of the cluster continuously
			for i := 0; i < numberOfCluster; i++ {
				client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(clusterConfigs[fmt.Sprintf("fake_cluster-%d", i)])
				for j := 0; j < 5; j++ {
					dep := fake_builder.BuildFakeDeployment(fmt.Sprintf("deployment-%d", j), "intuit.app", "app", "env", mockNamesapce)
					_, err := client.AppsV1().Deployments(mockNamesapce).Create(context.Background(), dep, metav1.CreateOptions{})
					Expect(err).To(BeNil())
				}
			}

			// Terminate all the controllers
			controller.StopAllControllers()

			// All the events picked should be processed completely
			// So the number of add events should be equal to the number of status events
			// Status events are triggered after add events are processed
			for i := 0; i < numberOfCluster; i++ {
				clusterName := fmt.Sprintf("fake_cluster-%d", i)
				Eventually(clusterHandlers[clusterName].AddCalled.Load, 30*time.Second).Should(Equal(clusterHandlers[clusterName].OnStatusCalled.Load()))
				Eventually(func() string {
					logMessageLock.Lock()
					defer logMessageLock.Unlock()
					return controllerLogMessages[clusterName][len(controllerLogMessages[clusterName])-1]
				}, 30*time.Second).Should(Equal("Process worker stopped"))
			}
		})
	})
})
