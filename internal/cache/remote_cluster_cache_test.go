package cache_test

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/fake/builder"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Remote Cluster Cache", Label("controllers_cache_test"), func() {
	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("add remote cluster to cache", func() {
		BeforeEach(func() {
			cache.RemoteCluster.Reset()
		})

		AfterEach(func() {
			cache.RemoteCluster.Reset()
		})

		It("should have all added remote clusters and pass relevant operations", func() {
			rc1 := builder.BuildRemoteCluster("cluster1")
			rc2 := builder.BuildRemoteCluster("cluster2")
			rc3 := builder.BuildRemoteCluster("cluster3")
			cache.RemoteCluster.AddCluster(rc1)
			cache.RemoteCluster.AddCluster(rc2)
			cache.RemoteCluster.AddCluster(rc3)
			foundrc1, _ := cache.RemoteCluster.GetCluster("cluster1")
			Expect(foundrc1).To(Equal(rc1))
			foundrc2, _ := cache.RemoteCluster.GetCluster("cluster2")
			Expect(foundrc2).To(Equal(rc2))
			foundrc3, _ := cache.RemoteCluster.GetCluster("cluster3")
			Expect(foundrc3).To(Equal(rc3))
			Expect(cache.RemoteCluster.GetCluster("cluster4")).To(BeNil())
			Expect(cache.RemoteCluster.ListClusters()).To(HaveLen(3))
			Expect(cache.RemoteCluster.ListClusters()).To(ContainElements(rc1, rc2, rc3))
			rc1 = builder.BuildRemoteCluster("cluster1new")
			cache.RemoteCluster.UpdateCluster("cluster1", rc1)
			foundrc1, _ = cache.RemoteCluster.GetCluster("cluster1")
			Expect(foundrc1).To(Equal(rc1))
		})
	})

	When("delete remote cluster from cache", func() {
		BeforeEach(func() {
			cache.RemoteCluster.Reset()
		})

		AfterEach(func() {
			cache.RemoteCluster.Reset()
		})

		It("should not have deleted remote cluster in cache", func() {
			rc1 := builder.BuildRemoteCluster("cluster1")
			rc2 := builder.BuildRemoteCluster("cluster2")
			rc3 := builder.BuildRemoteCluster("cluster3")
			cache.RemoteCluster.AddCluster(rc1)
			cache.RemoteCluster.AddCluster(rc2)
			cache.RemoteCluster.AddCluster(rc3)
			foundrc1, _ := cache.RemoteCluster.GetCluster("cluster1")
			Expect(foundrc1).To(Equal(rc1))
			foundrc2, _ := cache.RemoteCluster.GetCluster("cluster2")
			Expect(foundrc2).To(Equal(rc2))
			foundrc3, _ := cache.RemoteCluster.GetCluster("cluster3")
			Expect(foundrc3).To(Equal(rc3))
			Expect(cache.RemoteCluster.GetCluster("cluster4")).To(BeNil())
			Expect(cache.RemoteCluster.ListClusters()).To(HaveLen(3))
			Expect(cache.RemoteCluster.ListClusters()).To(ContainElements(rc1, rc2, rc3))
			cache.RemoteCluster.DeleteCluster("cluster1")
			Expect(cache.RemoteCluster.GetCluster("cluster1")).To(BeNil())
			Expect(cache.RemoteCluster.ListClusters()).To(HaveLen(2))
			Expect(cache.RemoteCluster.ListClusters()).To(ContainElements(rc2, rc3))
		})
	})
})
