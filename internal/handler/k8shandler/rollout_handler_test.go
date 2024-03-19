package k8shandler

import (
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

var _ = Describe("Test rollout handler operations", func() {
	var rolloutHandler handler.Handler
	var ctx context.Context
	var logMessages []string

	BeforeEach(func() {
		options.InitializeNaavikArgs(nil)
		cache.ResetAllCaches()
		rolloutHandler = NewRolloutHandler("cluster1", RolloutHandlerOpts{})
		ctx = context.NewContextWithLogger()
		logMessages = []string{}
		ctx.Log = ctx.Log.Hook(func(level string, msg string) {
			logMessages = append(logMessages, msg)
		})
	})

	AfterEach(func() {
		cache.ResetAllCaches()
	})

	Context("new rollout is added", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				rolloutHandler.Added(ctx, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Rollout object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("rollout workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "", "appName", "env", "namespace")
				rolloutHandler.Added(ctx, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid rollout added and cache not warmed up", func() {
			It("should update rollout to cache and not trigger handler", func() {
				options.StartUpTime = time.Now()
				options.InitializeNaavikArgs(&options.NaavikArgs{
					CacheRefreshInterval: 1 * time.Minute,
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "assetAlias", "appName", "env", "namespace")
				rolloutHandler.Added(ctx, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Adding rollout to cluster rollout cache"))
				Expect(logMessages).To(ContainElement("Cache not warmed up, skipping handling."))
				Expect(cache.Rollouts.GetByClusterIdentity("cluster1", "assetAlias").Rollouts["env"].Rollout).To(Equal(rollout))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid rollout added and cache warmed up", func() {
			It("should update rollout to cache and trigger handler", func() {
				options.StartUpTime = time.Now().Add(-5 * time.Minute)
				options.InitializeNaavikArgs(&options.NaavikArgs{
					CacheRefreshInterval: 1 * time.Second,
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "assetAlias", "appName", "env", "namespace")
				rolloutHandler.Added(ctx, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Adding rollout to cluster rollout cache"))
				Expect(logMessages).To(ContainElement("Triggering traffic config handler for identity started"))
				Expect(statusChan).To(BeClosed())
			})
		})
	})

	Context("rollout is updated", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				rolloutHandler.Updated(ctx, depcy, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Rollout object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("updated rollout workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "", "appName", "env", "namespace")
				rolloutHandler.Updated(ctx, rollout, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid rollout updated", func() {
			It("should update rollout to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "assetAlias", "appName", "env", "namespace")
				rolloutHandler.Updated(ctx, rollout, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier did not change, adding identity to cluster cache"))
				Expect(cache.Rollouts.GetByClusterIdentity("cluster1", "assetAlias").Rollouts["env"].Rollout).To(Equal(rollout))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid rollout updated with new assetAlias", func() {
			It("should update rollout to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "assetAlias", "appName", "env", "namespace")
				newDeploy := k8s_builder.BuildFakeRollout("rollout", "newAssetAlias", "appName", "env", "namespace")
				rolloutHandler.Updated(ctx, newDeploy, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier changed, adding cluster to new identity"))
				Expect(cache.Rollouts.GetByClusterIdentity("cluster1", "assetAlias")).To(BeNil())
				Expect(cache.Rollouts.GetByClusterIdentity("cluster1", "newAssetAlias").Rollouts["env"].Rollout).To(Equal(newDeploy))
				Expect(statusChan).To(BeClosed())
			})
		})
	})

	Context("rollout is deleted", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				rolloutHandler.Deleted(ctx, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting Rollout object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("deleted rollout workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "", "appName", "env", "namespace")
				rolloutHandler.Deleted(ctx, rollout, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid rollout deleted", func() {
			It("should update rollout to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				rollout := k8s_builder.BuildFakeRollout("rollout", "assetAlias", "appName", "env", "namespace")
				rolloutHandler.Deleted(ctx, rollout, statusChan)
				compelted := <-statusChan
				rolloutHandler.OnStatus(ctx, compelted)
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Rollout workload identifier deleted from rollouts and cluster cache"))
				Expect(cache.Rollouts.GetByClusterIdentity("cluster1", "assetAlias")).To(BeNil())
				Expect(statusChan).To(BeClosed())
			})
		})
	})
})
