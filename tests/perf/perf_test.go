package perf

import (
	goctx "context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	fake_builder "github.com/intuit/naavik/internal/fake/builder/resource"
	fake_controller "github.com/intuit/naavik/internal/fake/controller"
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	"github.com/intuit/naavik/internal/leasechecker"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	admiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	"github.com/jedib0t/go-pretty/v6/table"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var _ = Describe("Traffic config processing performance test", Label("traffic_config_perf"), func() {
	BeforeEach(func(sctx SpecContext) {
		cache.ResetAllCaches()
		logger.Log.SetLogLevel("info")
		args := &options.NaavikArgs{
			WorkerConcurrency:    1,
			LogLevel:             logger.DebugLevel.String(),
			CacheRefreshInterval: 5 * time.Second,
		}
		ctx := context.NewContextWithLogger()
		leaseChecker := leasechecker.GetStateChecker(ctx, types.StateCheckerNone)
		leaseChecker.RunStateCheck(ctx)
		options.InitializeNaavikArgs(args)
	})

	Context("Evaluate Naavik Performance", func() {
		It("Should arrive at desired resources and print the timetaken by various controllers in queue and for processing", func() {
			// Get the client for main cluster where naavik is running
			config, _ := fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath("main_cluster")
			client, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
			admiralClient, _ := fake_k8s_utils.NewFakeConfigLoader().AdmiralClientFromPath("main_cluster")

			// Data on which resources are created
			noOfClusters := 200
			noOfDeploymentNamespaces := 6
			noOfRolloutNamespaces := 5
			noOfDeploymentsInOneNamespace := 1
			noOfRolloutsInOneNamespace := 1
			noOfRolloutSvcInOneNamespace := 1
			noOfDeploymentSvcInOneNamespace := 1

			// Create remote cluster secrets based on noOfClusters
			fake_builder.CreateNRemoteClusterSecret(client, noOfClusters, "cluster", options.GetClusterRegistriesNamespace())

			// Start fake secret controllers, this will spin up fake remote clusters
			fake_controller.NewFakeSecretController(config, options.GetClusterRegistriesNamespace(), 10*time.Second)
			waitImmediate(func(ctx goctx.Context) (bool, error) {
				logger.Log.Infof("waiting for remote clusters to be added to cache. %d/%d", len(cache.RemoteCluster.ListClusters()), noOfClusters)
				return len(cache.RemoteCluster.ListClusters()) == noOfClusters, nil
			})

			// Create fake deployments, rollouts and there services in all the remote clusters
			for _, rc := range cache.RemoteCluster.ListClusters() {
				fake_builder.CreateNFakeDeploymentsInNNamespace(rc.K8sClient(), "deploy", "app-deploy", "deploy-namespace",
					rc.GetClusterID(), "qa", noOfDeploymentsInOneNamespace, noOfDeploymentNamespaces,
				)
				fake_builder.CreateNFakeServicesInNNamespace(rc.K8sClient(), "service", "app-deploy", "deploy-namespace",
					noOfDeploymentSvcInOneNamespace, noOfDeploymentNamespaces)
				fake_builder.CreateNFakeRolloutsInNNamespace(rc.ArgoClient(), "rollout", "app-rollout", "rollout-namespace",
					rc.GetClusterID(), "qa", noOfRolloutsInOneNamespace, noOfRolloutNamespaces,
				)
				fake_builder.CreateNFakeServicesInNNamespace(rc.K8sClient(), "service", "app-rollout", "rollout-namespace",
					noOfRolloutSvcInOneNamespace, noOfRolloutNamespaces)
			}

			// Create dependency network
			// Service in cluster-N namespace-N can talk to service in cluster-2 namespace-N+1 to namespace-N+maxDependencyNetwork
			maxDependencyNetwork := 3
			fake_builder.CreateNFakeDependencies(admiralClient, noOfClusters, noOfDeploymentNamespaces, "deploy-namespace", "deploy-1", maxDependencyNetwork)
			fake_builder.CreateNFakeDependencies(admiralClient, noOfClusters, noOfRolloutNamespaces, "rollout-namespace", "rollout-1", maxDependencyNetwork)

			// Start the dependency controller
			fake_controller.NewFakeDependencyController(config, options.GetDependenciesNamespace(), 0)
			// Start the traffic config controller
			fake_controller.NewFakeTrafficConfigController(config, options.GetTrafficConfigNamespace(), 0)

			// Wait for all the deployments, rollouts, service, dependency to be added to cache
			waitImmediate(func(ctx goctx.Context) (bool, error) {
				logger.Log.Infof(
					"caches are getting synced, clusters=%d/%d deployments=%d/%d rollouts=%d/%d depedencyrecord=%d/%d",
					len(cache.RemoteCluster.ListClusters()), noOfClusters,
					cache.Deployments.GetNoOfDeployments(), noOfClusters*noOfDeploymentNamespaces*noOfDeploymentsInOneNamespace,
					cache.Rollouts.GetNoOfRollouts(), noOfClusters*noOfRolloutNamespaces*noOfRolloutsInOneNamespace,
					cache.IdentityDependency.GetTotalDependencies(), noOfClusters*noOfDeploymentNamespaces+noOfClusters*noOfRolloutNamespaces,
					cache.TrafficConfigCache.GetTotalTrafficConfigs(), noOfClusters*(noOfDeploymentNamespaces+noOfRolloutNamespaces),
				)
				return cache.Deployments.GetNoOfDeployments() == noOfClusters*noOfDeploymentNamespaces*noOfDeploymentsInOneNamespace &&
					cache.Rollouts.GetNoOfRollouts() == noOfClusters*noOfRolloutNamespaces*noOfRolloutsInOneNamespace &&
					cache.IdentityDependency.GetTotalDependencies() == noOfClusters*noOfDeploymentNamespaces+noOfClusters*noOfRolloutNamespaces, nil
			})
			logger.Log.Info("All deployments, rollouts and dependencies are added to cache")
			printDependencyGraph()

			// Wait till cache warm up completes
			waitImmediate(func(ctx goctx.Context) (bool, error) {
				logger.Log.Infof("waiting for cache to warm up")
				return options.IsCacheWarmedUp(), nil
			})

			updateAllTrafficConfigs(admiralClient, noOfClusters, noOfDeploymentNamespaces, noOfRolloutNamespaces)

			time.Sleep(15 * time.Second)

			// Validate the number of envoyfilters (throttle and router) and virtualservices created in all the clusters
			// TODO: Try to validate the spec of envoyfilters and virtualservices
			for _, cluster := range cache.RemoteCluster.ListClusters() {
				// Get all the envoyfilters in the cluster
				throttleEnvoyFilters, _ := cluster.IstioClient().ListEnvoyFilters(context.NewContextWithLogger(), types.NamespaceIstioSystem, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=throttle_filter", types.CreatedTypeKey),
				})
				Expect(len(throttleEnvoyFilters.Items)).To(Equal(noOfDeploymentNamespaces+noOfRolloutNamespaces), cluster.GetClusterID())

				routerEnvoyFilters, _ := cluster.IstioClient().ListEnvoyFilters(context.NewContextWithLogger(), types.NamespaceIstioSystem, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("%s=router_filter", types.CreatedTypeKey),
				})
				if !strings.HasSuffix(cluster.GetClusterID(), fmt.Sprint(noOfClusters)) {
					Expect(len(routerEnvoyFilters.Items)).To(Equal((noOfDeploymentNamespaces)+(noOfRolloutNamespaces)-2), cluster.GetClusterID())
				} else {
					Expect(len(routerEnvoyFilters.Items)).To(Equal(0))
				}

				virtualservices, _ := cluster.IstioClient().ListVirtualServices(context.NewContextWithLogger(), options.GetSyncNamespace(), metav1.ListOptions{})
				if !strings.HasSuffix(cluster.GetClusterID(), fmt.Sprint(noOfClusters)) {
					Expect(len(virtualservices.Items)).To(Equal((noOfDeploymentNamespaces)+(noOfRolloutNamespaces)-2), cluster.GetClusterID())
				} else {
					Expect(len(virtualservices.Items)).To(Equal(0))
				}
			}
		})
	})
})

func updateAllTrafficConfigs(admiralClient admiralclientset.Interface, noOfClusters, noOfDeploymentNamespaces, noOfRolloutNamespaces int) {
	// Create traffic config object for all the deployments, rollouts across all the clusters
	fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
		admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfDeploymentNamespaces,
		"cluster", "deploy-namespace", "deploy-1", "qa", "2",
	)
	fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
		admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfRolloutNamespaces,
		"cluster", "rollout-namespace", "rollout-1", "qa", "2",
	)
}

func concurrentlyUpdateTrafficConfig(admiralClient admiralclientset.Interface, noOfClusters, noOfDeploymentNamespaces, noOfRolloutNamespaces int) {
	go func() {
		// Create traffic config object for all the deployments, rollouts across all the clusters
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfDeploymentNamespaces,
			"cluster", "deploy-namespace", "deploy-1", "qa", "2",
		)
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfRolloutNamespaces,
			"cluster", "rollout-namespace", "rollout-1", "qa", "2",
		)
	}()

	go func() {
		// Create traffic config object for all the deployments, rollouts across all the clusters
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfDeploymentNamespaces,
			"cluster", "deploy-namespace", "deploy-1", "qa", "3",
		)
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfRolloutNamespaces,
			"cluster", "rollout-namespace", "rollout-1", "qa", "3",
		)
	}()

	go func() {
		// Create traffic config object for all the deployments, rollouts across all the clusters
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfDeploymentNamespaces,
			"cluster", "deploy-namespace", "deploy-1", "qa", "4",
		)
		fake_builder.CreateFakeTrafficConfigsInNNamespaceNClusters(
			admiralClient, options.GetTrafficConfigNamespace(), noOfClusters, noOfRolloutNamespaces,
			"cluster", "rollout-namespace", "rollout-1", "qa", "4",
		)
	}()
}

func waitImmediate(waitFunc wait.ConditionWithContextFunc) {
	wait.PollUntilContextTimeout(goctx.Background(), 1*time.Second, 2*time.Minute, true, waitFunc)
}

func printDependencyGraph() {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Dependency Graph Generated"})
	t.AppendHeader(table.Row{"Source", "Destinations"})
	i := 0
	cache.IdentityDependency.RangedDependencies(func(identity string, dependencies []string) bool {
		t.AppendRow(table.Row{identity, dependencies})
		t.AppendSeparator()
		i++
		return true
	})
	t.Render()
}
