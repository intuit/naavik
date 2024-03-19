package cache_test

import (
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	cache_builder "github.com/intuit/naavik/internal/fake/builder/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Services Cache", Label("services_cache_test"), func() {
	var noOfClusters int = 100
	var noOfNamespaces int = 50
	var noOfservicesInNamespace int = 4

	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("services are added to cache", func() {
		BeforeEach(func() {
			resourcebuildmap := cache_builder.CreateClusterNamespaceNumResourcesMap(noOfClusters, noOfNamespaces, noOfservicesInNamespace)
			cache_builder.BuildFakeServiceCache(resourcebuildmap)
		})

		AfterEach(func() {
			cache.Services.Reset()
		})

		It("should have all services in cache", func() {
			for i := 0; i < noOfClusters; i++ {
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					for k := 0; k < noOfservicesInNamespace; k++ {
						namespaceName := fmt.Sprintf("namespace-%d", j)
						serviceEntry := cache.Services.GetByClusterNamespace(clusterName, namespaceName)
						servicesByIdentity := cache.Services.GetByNamespace(namespaceName)
						Expect(len(serviceEntry.Services)).To(Equal(noOfservicesInNamespace))
						Expect(len(servicesByIdentity)).To(Equal(noOfClusters * noOfservicesInNamespace))
					}
				}
			}
		})

		It("should delete service one by one from cache", func() {
			for i := 0; i < noOfClusters; i++ {
				identityServiceCount := map[string]int{}
				clusterName := fmt.Sprintf("cluster-%d", i)
				for j := 0; j < noOfNamespaces; j++ {
					namespaceName := fmt.Sprintf("namespace-%d", j)
					_, ok := identityServiceCount[namespaceName]
					if !ok {
						identityServiceCount[namespaceName] = noOfservicesInNamespace
					}
					namespaceServices := cache.Services.GetByClusterNamespace(clusterName, namespaceName)
					Expect(len(namespaceServices.Services)).To(Equal(identityServiceCount[namespaceName]), namespaceName)
					for _, svc := range namespaceServices.Services {
						cache.Services.Delete(clusterName, svc.Service)
						identityServiceCount[namespaceName] -= 1
					}
					latestSvcs := cache.Services.GetByClusterNamespace(clusterName, namespaceName)
					Expect(latestSvcs).To(BeNil())
				}
			}
		})
	})
})
