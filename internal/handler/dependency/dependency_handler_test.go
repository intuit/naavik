package dependencyhandler

import (
	"testing"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	k8s_builder "github.com/intuit/naavik/internal/fake/builder/resource"
	"github.com/intuit/naavik/internal/handler"
	"github.com/intuit/naavik/internal/types/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDependencyHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dependency_handler_test")
}

var _ = Describe("Test dependency handler operations", func() {
	var dependenyHandler handler.Handler
	var ctx context.Context
	var logMessages []string

	BeforeEach(func() {
		cache.ResetAllCaches()
		dependenyHandler = NewDependencyHandler(Opts{})
		ctx = context.NewContextWithLogger()
		logMessages = []string{}
		ctx.Log = ctx.Log.Hook(func(level string, msg string) {
			logMessages = append(logMessages, msg)
		})
	})

	AfterEach(func() {
		cache.ResetAllCaches()
	})

	Context("new dependency is added", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deployment := k8s_builder.BuildFakeDeployment("", "", "", "", "")
				dependenyHandler.Added(ctx, deployment, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Dependency object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})
		When("dependency is added with no source", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{"destination1", "destination2", "destination3", "destination4"})
				dependenyHandler.Added(ctx, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error Dependency has no source, skipping handling."))
				Expect(logMessages).NotTo(ContainElement("add to dependency cache"))
				Expect(statusChan).To(BeClosed())
			})
		})

		It("should update the dependencies to cache and channel should be closed", func() {
			statusChan := make(chan controller.EventProcessStatus, 1)
			depcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3", "destination4"})
			dependenyHandler.Added(ctx, depcy, statusChan)
			compelted := <-statusChan
			Expect(compelted).NotTo(BeNil())
			Expect(statusChan).To(BeClosed())
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("source")).To(HaveLen(4))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("destination1")).To(HaveLen(1))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("Destination2")).To(HaveLen(1))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("desTination3")).To(HaveLen(1))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("destination4")).To(HaveLen(1))
			Expect(logMessages).To(ContainElement("add to dependency cache"))
		})
	})

	Context("new dependency is updated", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deployment := k8s_builder.BuildFakeDeployment("", "", "", "", "")
				dependenyHandler.Updated(ctx, deployment, deployment, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Dependency object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})
		When("dependency is updated with no source", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{"destination1", "destination2", "destination3", "destination4"})
				dependenyHandler.Updated(ctx, depcy, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error Dependency has no source, skipping handling."))
				Expect(logMessages).NotTo(ContainElement("update to dependency cache"))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("dependency is updated with source and destinations", func() {
			It("should update the dependencies to cache and channel should be closed", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3", "destination4"})
				dependenyHandler.Updated(ctx, depcy, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(statusChan).To(BeClosed())
				Expect(cache.IdentityDependency.GetDependenciesForIdentity("source")).To(HaveLen(4))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination1")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("Destination2")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("desTination3")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination4")).To(HaveLen(1))
				stash := []string{}
				Expect(logMessages).To(ContainElement("update to dependency cache", &stash))
				Expect(stash).To(HaveLen(4))
			})
		})

		When("dependency is updated with new destination", func() {
			It("should update new dependency to cache and trigger handlers", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				olddepcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3"})
				newDepcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3", "destination4", "destination5"})
				dependenyHandler.Updated(ctx, newDepcy, olddepcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(statusChan).To(BeClosed())
				Expect(cache.IdentityDependency.GetDependenciesForIdentity("source")).To(HaveLen(5))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination1")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("Destination2")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("desTination3")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination4")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination5")).To(HaveLen(1))
				stash := []string{}
				Expect(logMessages).To(ContainElement("update to dependency cache", &stash))
				Expect(stash).To(HaveLen(5))
				// Traffic config handler should be triggered for new destination
				stash = []string{}
				Expect(logMessages).To(ContainElement("New destination found, triggering handlers", &stash))
				Expect(stash).To(HaveLen(2))
			})
		})

		When("cache not warmed up", func() {
			It("should update new dependency to cache and not trigger handlers", func() {
				// Set the startup time to 1 minute ago and cache refresh interval to 5 minutes
				options.StartUpTime = time.Now().Add(time.Minute * -1)
				options.InitializeNaavikArgs(&options.NaavikArgs{
					CacheRefreshInterval: 5 * time.Minute,
				})

				statusChan := make(chan controller.EventProcessStatus, 1)
				olddepcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3"})
				newDepcy := k8s_builder.BuildFakeDependency("namespace", "soUrce", []string{"destination1", "destination2", "destination3", "destination4", "destination5"})
				dependenyHandler.Updated(ctx, newDepcy, olddepcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(statusChan).To(BeClosed())
				Expect(cache.IdentityDependency.GetDependenciesForIdentity("source")).To(HaveLen(5))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination1")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("Destination2")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("desTination3")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination4")).To(HaveLen(1))
				Expect(cache.IdentityDependency.GetDependentsForIdentity("destination5")).To(HaveLen(1))
				stash := []string{}
				Expect(logMessages).To(ContainElement("update to dependency cache", &stash))
				Expect(stash).To(HaveLen(5))
				// Traffic config handler should not be triggered for new destination
				Expect(logMessages).NotTo(ContainElement("NoOp TrafficConfig Trigger"))
			})
		})
	})

	Context("dependency is deleted", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deployment := k8s_builder.BuildFakeDeployment("", "", "", "", "")
				dependenyHandler.Deleted(ctx, deployment, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Dependency object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})
		When("dependency is deleted", func() {
			It("should not do anything and simply close channel", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{"destination1", "destination2", "destination3", "destination4"})
				dependenyHandler.Deleted(ctx, depcy, statusChan)
				compelted := <-statusChan
				dependenyHandler.OnStatus(ctx, compelted)
				Expect(compelted).NotTo(BeNil())
				Expect(statusChan).To(BeClosed())
			})
		})
	})
})
