package cache_test

import (
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	cache_builder "github.com/intuit/naavik/internal/fake/builder/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Rollouts Cache", Label("rollouts_cache_test"), func() {
	var noOfClusters int = 100
	var noOfNamespaces int = 50
	var noOfrolloutsInNamespace int = 4

	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("rollouts are added to cache", func() {
		BeforeEach(func() {
			cache.Rollouts.Reset()
			resourcebuildmap := cache_builder.CreateClusterNamespaceNumResourcesMap(noOfClusters, noOfNamespaces, noOfrolloutsInNamespace)
			cache_builder.BuildFakeRolloutsCache(resourcebuildmap)
		})
		AfterEach(func() {
			cache.Rollouts.Reset()
		})
		It("should have all rollouts in cache", func() {
			for i := 0; i < noOfClusters; i++ {
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					for k := 0; k < noOfrolloutsInNamespace; k++ {
						env := fmt.Sprintf("env-%d", k)
						namespaceName := fmt.Sprintf("intuit.namespace-%d.rollout", j)
						rolloutEntry := cache.Rollouts.GetByClusterIdentity(clusterName, namespaceName)
						rolloutEntryByEnv := cache.Rollouts.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						rolloutsByIdentity := cache.Rollouts.GetByIdentity(namespaceName)
						Expect(rolloutEntryByEnv).NotTo(BeNil())
						Expect(len(rolloutEntry.Rollouts)).To(Equal(noOfrolloutsInNamespace))
						Expect(len(rolloutsByIdentity)).To(Equal(noOfClusters * noOfrolloutsInNamespace))
					}
				}
			}
			Expect(cache.Rollouts.GetNoOfRollouts()).To(Equal(noOfClusters * noOfNamespaces * noOfrolloutsInNamespace))
		})

		It("should delete rollout one by one from cache", func() {
			identityRolloutCount := map[string]int{}
			for i := 0; i < noOfClusters; i++ {
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					namespaceName := fmt.Sprintf("intuit.namespace-%d.rollout", j)
					_, ok := identityRolloutCount[namespaceName]
					if !ok {
						identityRolloutCount[namespaceName] = noOfClusters * noOfrolloutsInNamespace
					}
					for k := 0; k < noOfrolloutsInNamespace; k++ {
						env := fmt.Sprintf("env-%d", k)
						rollout := cache.Rollouts.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						cache.Rollouts.Delete(clusterName, rollout)
						identityRolloutCount[namespaceName] = identityRolloutCount[namespaceName] - 1
						rolloutsByIdentity := cache.Rollouts.GetByIdentity(namespaceName)
						rolloutEntryByEnv := cache.Rollouts.GetByClusterIdentityEnv(clusterName, namespaceName, env)
						Expect(len(rolloutsByIdentity)).To(Equal(identityRolloutCount[namespaceName]), namespaceName)
						Expect(rolloutEntryByEnv).To(BeNil())
					}
					rolloutEntry := cache.Rollouts.GetByClusterIdentity(clusterName, namespaceName)
					Expect(rolloutEntry).To(BeNil())
				}
			}
		})
	})
})
