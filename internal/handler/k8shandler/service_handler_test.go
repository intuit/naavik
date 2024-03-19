package k8shandler

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	k8s_builder "github.com/intuit/naavik/internal/fake/builder/resource"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test service handler operations", func() {
	var serviceHandler handler.Handler
	var ctx context.Context
	var logMessages []string

	BeforeEach(func() {
		options.InitializeNaavikArgs(nil)
		cache.ResetAllCaches()
		serviceHandler = NewServiceHandler("cluster1", ServiceHandlerOpts{})
		ctx = context.NewContextWithLogger()
		logMessages = []string{}
		ctx.Log = ctx.Log.Hook(func(level string, msg string) {
			logMessages = append(logMessages, msg)
		})
	})

	AfterEach(func() {
		cache.ResetAllCaches()
	})

	Context("new service is added", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				serviceHandler.Added(ctx, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Service object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})

			It("should add service to the cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeService("service", "assetAlias", "app", "namespace")
				serviceHandler.Added(ctx, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Adding service to cluster service cache"))
				Expect(statusChan).To(BeClosed())
				Expect(cache.Services.GetByClusterNamespace("cluster1", "namespace")).NotTo(BeNil())
				Expect(cache.Services.GetByClusterNamespace("cluster1", "namespace").Services["service"].Service).To(Equal(service))
			})
		})
	})

	Context("service is updated", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				serviceHandler.Updated(ctx, service, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Service object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})

			It("should add service to the cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeService("service", "assetAlias", "app", "namespace")
				serviceHandler.Updated(ctx, service, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Updating service to cluster service cache"))
				Expect(statusChan).To(BeClosed())
				Expect(cache.Services.GetByClusterNamespace("cluster1", "namespace")).NotTo(BeNil())
				Expect(cache.Services.GetByClusterNamespace("cluster1", "namespace").Services["service"].Service).To(Equal(service))
			})
		})
	})

	Context("service is deleted", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				serviceHandler.Deleted(ctx, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Service object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})

			It("should delete service from the cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				service := k8s_builder.BuildFakeService("service", "assetAlias", "app", "namespace")
				serviceHandler.Updated(ctx, service, service, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Updating service to cluster service cache"))
				Expect(statusChan).To(BeClosed())

				logMessages = []string{}
				statusChan = make(chan controller.EventProcessStatus, 1)
				serviceHandler.Deleted(ctx, service, statusChan)
				compelted = <-statusChan
				serviceHandler.OnStatus(ctx, compelted)
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deleting service from cluster service cache"))
				Expect(statusChan).To(BeClosed())
			})
		})
	})
})
