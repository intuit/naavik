package cache_test

import (
	"fmt"
	"sync"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	cache_builder "github.com/intuit/naavik/internal/fake/builder/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Deployment Cache", Label("deployments_cache_test"), func() {
	var noOfClusters int = 100
	var noOfNamespaces int = 50
	var noOfDeploymentsInNamespace int = 4

	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	AfterEach(func() {
		cache.Deployments.Reset()
	})

	When("Deployments are added to cache", func() {
		BeforeEach(func() {
			resourcebuildmap := cache_builder.CreateClusterNamespaceNumResourcesMap(noOfClusters, noOfNamespaces, noOfDeploymentsInNamespace)
			cache_builder.BuildFakeDeploymentCache(resourcebuildmap)
		})

		It("should add all deployments to cache", func() {
			for i := 0; i < noOfClusters; i++ {
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					for k := 0; k < noOfDeploymentsInNamespace; k++ {
						env := fmt.Sprintf("env-%d", k)
						namespaceName := fmt.Sprintf("intuit.namespace-%d.deploy", j)
						depEntry := cache.Deployments.GetByClusterIdentity(clusterName, namespaceName)
						depEntryByEnv := cache.Deployments.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						depsByIdentity := cache.Deployments.GetByIdentity(namespaceName)
						Expect(depEntryByEnv).NotTo(BeNil())
						Expect(len(depEntry.Deployments)).To(Equal(noOfDeploymentsInNamespace))
						Expect(len(depsByIdentity)).To(Equal(noOfClusters * noOfDeploymentsInNamespace))
					}
				}
			}
			Expect(cache.Deployments.GetNoOfDeployments()).To(Equal(noOfClusters * noOfNamespaces * noOfDeploymentsInNamespace))
		})

		It("should delete deployment one by one from cache", func() {
			identityDepCount := map[string]int{}
			for i := 0; i < noOfClusters; i++ {
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					namespaceName := fmt.Sprintf("intuit.namespace-%d.deploy", j)
					_, ok := identityDepCount[namespaceName]
					if !ok {
						identityDepCount[namespaceName] = noOfClusters * noOfDeploymentsInNamespace
					}
					var wg sync.WaitGroup
					for k := 0; k < noOfDeploymentsInNamespace; k++ {
						env := fmt.Sprintf("env-%d", k)
						deploy := cache.Deployments.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						wg.Add(1)
						func() {
							defer wg.Done()
							cache.Deployments.Delete(clusterName, deploy)
						}()
						wg.Wait()
						identityDepCount[namespaceName] = identityDepCount[namespaceName] - 1
						depsByIdentity := cache.Deployments.GetByIdentity(namespaceName)
						depEntryByEnv := cache.Deployments.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						Expect(len(depsByIdentity)).To(Equal(identityDepCount[namespaceName]), namespaceName)
						Expect(depEntryByEnv).To(BeNil())
					}
					depEntry := cache.Deployments.GetByClusterIdentity(clusterName, namespaceName)
					Expect(depEntry).To(BeNil())
				}
			}
		})
	})
})
