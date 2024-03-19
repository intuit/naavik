package cache_test

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Identity Cluster Cache", Label("identity_cluster_cache_test"), func() {
	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("cluster is added for an identity", func() {
		BeforeEach(func() {
			cache.IdentityCluster.Reset()
		})

		AfterEach(func() {
			cache.IdentityCluster.Reset()
		})

		It("should add the cluster has part of identity and relevant operations should succeed", func() {
			cache.IdentityCluster.AddClusterToIdentity("identity-1", "cluster-1")
			cache.IdentityCluster.AddClusterToIdentity("identity-1", "cluster-2")
			cache.IdentityCluster.AddClusterToIdentity("identity-2", "cluster-1")
			cache.IdentityCluster.AddClusterToIdentity("identity-3", "cluster-1")
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-notpresent")).To(HaveLen(0))
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).To(ContainElements("cluster-1", "cluster-2"))
			Expect(cache.IdentityCluster.ListIdentities()).To(HaveLen(3))
			Expect(cache.IdentityCluster.ListIdentities()).To(ContainElements("identity-1", "identity-2", "identity-3"))
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-1")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-2")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-3")).To(BeFalse())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-2", "cluster-1")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-3", "cluster-1")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-notpresent", "cluster-1")).To(BeFalse())
		})
	})

	When("cluster is deleted for identity", func() {
		BeforeEach(func() {
			cache.IdentityCluster.Reset()
		})

		AfterEach(func() {
			cache.IdentityCluster.Reset()
		})

		It("should delete the cluster from identity and relevant operations should succeed", func() {
			cache.IdentityCluster.AddClusterToIdentity("identity-1", "cluster-1")
			cache.IdentityCluster.AddClusterToIdentity("identity-1", "cluster-2")
			cache.IdentityCluster.AddClusterToIdentity("identity-2", "cluster-1")
			cache.IdentityCluster.AddClusterToIdentity("identity-3", "cluster-1")
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).To(ContainElements("cluster-1", "cluster-2"))
			Expect(cache.IdentityCluster.ListIdentities()).To(HaveLen(3))
			Expect(cache.IdentityCluster.ListIdentities()).To(ContainElements("identity-1", "identity-2", "identity-3"))
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-1")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-2")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-1", "cluster-3")).To(BeFalse())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-2", "cluster-1")).To(BeTrue())
			Expect(cache.IdentityCluster.IsClusterPresentInIdentity("identity-3", "cluster-1")).To(BeTrue())
			cache.IdentityCluster.DeleteClusterFromIdentity("identity-1", "cluster-1")
			cache.IdentityCluster.DeleteClusterFromIdentity("identity-1", "cluster-2")
			cache.IdentityCluster.DeleteClusterFromIdentity("identity-notpresent", "cluster-2")
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).To(HaveLen(0))
			Expect(cache.IdentityCluster.GetClustersForIdentity("identity-1")).ToNot(ContainElements("cluster-1", "cluster-2"))
			Expect(cache.IdentityCluster.ListIdentities()).To(HaveLen(3))
		})
	})
})
