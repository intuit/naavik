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

var _ = Describe("Test deployment handler operations", func() {
	var deploymentHandler handler.Handler
	var ctx context.Context
	var logMessages []string

	BeforeEach(func() {
		options.InitializeNaavikArgs(nil)
		cache.ResetAllCaches()
		deploymentHandler = NewDeploymentHandler("cluster1", DeploymentHandlerOpts{})
		ctx = context.NewContextWithLogger()
		logMessages = []string{}
		ctx.Log = ctx.Log.Hook(func(level string, msg string) {
			logMessages = append(logMessages, msg)
		})
	})

	AfterEach(func() {
		cache.ResetAllCaches()
	})

	Context("new deployment is added", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				deploymentHandler.Added(ctx, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting deployment object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("deployment workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "", "appName", "env", "namespace")
				deploymentHandler.Added(ctx, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid deployment added and cache not warmed up", func() {
			It("should update deployment to cache and not trigger handler", func() {
				options.StartUpTime = time.Now()
				options.InitializeNaavikArgs(&options.NaavikArgs{
					CacheRefreshInterval: 1 * time.Minute,
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "assetAlias", "appName", "env", "namespace")
				deploymentHandler.Added(ctx, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Adding deploy to cluster deployment cache"))
				Expect(logMessages).To(ContainElement("Cache not warmed up, skipping handling."))
				Expect(cache.Deployments.GetByClusterIdentity("cluster1", "assetAlias").Deployments["env"].Deployment).To(Equal(deploy))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid deployment added and cache warmed up", func() {
			It("should update deployment to cache and trigger handler", func() {
				options.StartUpTime = time.Now().Add(-5 * time.Minute)
				options.InitializeNaavikArgs(&options.NaavikArgs{
					CacheRefreshInterval: 1 * time.Second,
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "assetAlias", "appName", "env", "namespace")
				deploymentHandler.Added(ctx, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Adding deploy to cluster deployment cache"))
				Expect(logMessages).To(ContainElement("Triggering traffic config handler for identity started"))
				Expect(statusChan).To(BeClosed())
			})
		})
	})

	Context("deployment is updated", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				deploymentHandler.Updated(ctx, depcy, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting deployment object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("updated deployment workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "", "appName", "env", "namespace")
				deploymentHandler.Updated(ctx, deploy, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid deployment updated", func() {
			It("should update deployment to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "assetAlias", "appName", "env", "namespace")
				deploymentHandler.Updated(ctx, deploy, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier did not change, adding identity to cluster cache"))
				Expect(cache.Deployments.GetByClusterIdentity("cluster1", "assetAlias").Deployments["env"].Deployment).To(Equal(deploy))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid deployment updated with new assetAlias", func() {
			It("should update deployment to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "assetAlias", "appName", "env", "namespace")
				newDeploy := k8s_builder.BuildFakeDeployment("deployment", "newAssetAlias", "appName", "env", "namespace")
				deploymentHandler.Updated(ctx, newDeploy, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier changed, adding cluster to new identity"))
				Expect(cache.Deployments.GetByClusterIdentity("cluster1", "assetAlias")).To(BeNil())
				Expect(cache.Deployments.GetByClusterIdentity("cluster1", "newAssetAlias").Deployments["env"].Deployment).To(Equal(newDeploy))
				Expect(statusChan).To(BeClosed())
			})
		})
	})

	Context("deployment is deleted", func() {
		When("unable to cast received object", func() {
			It("should log error and do nothing", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				depcy := k8s_builder.BuildFakeDependency("namespace", "", []string{""})
				deploymentHandler.Deleted(ctx, depcy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("error casting deployment object, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("deleted deployment workload identifier is empty", func() {
			It("should skip handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "", "appName", "env", "namespace")
				deploymentHandler.Deleted(ctx, deploy, statusChan)
				compelted := <-statusChan
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier is empty, skipping handling."))
				Expect(statusChan).To(BeClosed())
			})
		})

		When("valid deployment deleted", func() {
			It("should update deployment to cache", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				deploy := k8s_builder.BuildFakeDeployment("deployment", "assetAlias", "appName", "env", "namespace")
				deploymentHandler.Deleted(ctx, deploy, statusChan)
				compelted := <-statusChan
				deploymentHandler.OnStatus(ctx, compelted)
				Expect(compelted).NotTo(BeNil())
				Expect(logMessages).To(ContainElement("Deployment workload identifier deleted from deployments and cluster cache"))
				Expect(cache.Deployments.GetByClusterIdentity("cluster1", "assetAlias")).To(BeNil())
				Expect(statusChan).To(BeClosed())
			})
		})
	})
})
