package cache_test

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Identity Dependency Cache", Label("identity_dependency_cache_test"), func() {
	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("dependency is added for an identity", func() {
		BeforeEach(func() {
			cache.IdentityDependency.Reset()
		})

		AfterEach(func() {
			cache.IdentityDependency.Reset()
		})

		It("should add the dependency for identity", func() {
			cache.IdentityDependency.AddDependencyToIdentity("identity-1", "dependency-1")
			cache.IdentityDependency.AddDependencyToIdentity("identity-1", "dependency-2")
			cache.IdentityDependency.AddDependencyToIdentity("identity-2", "dependency-1")
			cache.IdentityDependency.AddDependencyToIdentity("identity-3", "dependency-1")
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-notpresent")).To(HaveLen(0))
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-1", "dependency-1")).To(BeTrue())
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-1", "dependency-2")).To(BeTrue())
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-1", "dependency-3")).To(BeFalse())
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-2", "dependency-1")).To(BeTrue())
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-3", "dependency-1")).To(BeTrue())
			Expect(cache.IdentityDependency.IsDependentForIdentity("identity-notpresent", "dependency-1")).To(BeFalse())
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(ContainElements("dependency-1", "dependency-2"))
			Expect(cache.IdentityDependency.GetTotalDependencies()).To(Equal(3))
			cache.IdentityDependency.RangedDependencies(func(identity string, dependencies []string) bool {
				Expect(dependencies).To(ContainElement("dependency-1"))
				return true
			})
		})
	})

	When("dependent is added for an identity", func() {
		BeforeEach(func() {
			cache.IdentityDependency.Reset()
		})

		AfterEach(func() {
			cache.IdentityDependency.Reset()
		})

		It("should add the dependent for identity", func() {
			cache.IdentityDependency.AddDependentToIdentity("identity-1", "dependent-1")
			cache.IdentityDependency.AddDependentToIdentity("identity-1", "dependent-2")
			cache.IdentityDependency.AddDependentToIdentity("identity-2", "dependent-1")
			cache.IdentityDependency.AddDependentToIdentity("identity-3", "dependent-1")
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-notpresent")).To(HaveLen(0))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(ContainElements("dependent-1", "dependent-2"))
		})
	})

	When("dependency is deleted for an identity", func() {
		BeforeEach(func() {
			cache.IdentityCluster.Reset()
		})

		AfterEach(func() {
			cache.IdentityCluster.Reset()
		})

		It("should delete the dependency for identity and relevant operations should succeed", func() {
			cache.IdentityDependency.AddDependencyToIdentity("identity-1", "dependency-1")
			cache.IdentityDependency.AddDependencyToIdentity("identity-1", "dependency-2")
			cache.IdentityDependency.AddDependencyToIdentity("identity-2", "dependency-1")
			cache.IdentityDependency.AddDependencyToIdentity("identity-3", "dependency-1")
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(ContainElements("dependency-1", "dependency-2"))
			Expect(cache.IdentityDependency.GetTotalDependencies()).To(Equal(3))
			cache.IdentityDependency.DeleteDependencyFromIdentity("identity-1", "dependency-1")
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(HaveLen(1))
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(ContainElement("dependency-2"))
			Expect(cache.IdentityDependency.GetTotalDependencies()).To(Equal(3))
			cache.IdentityDependency.DeleteDependencyFromIdentity("identity-1", "dependency-2")
			Expect(cache.IdentityDependency.GetDependenciesForIdentity("identity-1")).To(HaveLen(0))
			Expect(cache.IdentityDependency.GetTotalDependencies()).To(Equal(2))
			cache.IdentityDependency.DeleteDependencyFromIdentity("identity-not-present", "dependency-2")
		})
	})

	When("dependenct is deleted for an identity", func() {
		BeforeEach(func() {
			cache.IdentityCluster.Reset()
		})

		AfterEach(func() {
			cache.IdentityCluster.Reset()
		})

		It("should delete the dependent for identity and relevant operations should succeed", func() {
			cache.IdentityDependency.AddDependentToIdentity("identity-1", "dependent-1")
			cache.IdentityDependency.AddDependentToIdentity("identity-1", "dependent-2")
			cache.IdentityDependency.AddDependentToIdentity("identity-2", "dependent-1")
			cache.IdentityDependency.AddDependentToIdentity("identity-3", "dependent-1")
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(HaveLen(2))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(ContainElements("dependent-1", "dependent-2"))
			cache.IdentityDependency.DeleteDependentFromIdentity("identity-1", "dependent-1")
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(HaveLen(1))
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(ContainElement("dependent-2"))
			cache.IdentityDependency.DeleteDependentFromIdentity("identity-1", "dependent-2")
			Expect(cache.IdentityDependency.GetDependentsForIdentity("identity-1")).To(HaveLen(0))
			cache.IdentityDependency.DeleteDependentFromIdentity("identity-not-present", "dependent-2")
		})
	})
})
