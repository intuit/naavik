package cache_test

import (
	"fmt"
	"sync"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	resourcebuilder "github.com/intuit/naavik/internal/fake/builder/resource"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test TrafficConfig Cache", Label("trafficconfig_cache_test"), func() {
	var trafficconfig *admiralv1.TrafficConfig
	noOfTrafficConfigs := 100
	noOfTrafficConfigEnvs := 10

	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("trafficconfigs are added to cache", func() {
		BeforeEach(func(sctx SpecContext) {
			var wg sync.WaitGroup
			for i := 0; i < noOfTrafficConfigs; i++ {
				for j := 0; j < noOfTrafficConfigEnvs; j++ {
					wg.Add(1)
					trafficconfig = resourcebuilder.GetFakeTrafficConfig("app.exampleasset-"+fmt.Sprintf("%d", i), fmt.Sprintf("env-%d", j), "1", "namespace")
					func() {
						defer wg.Done()
						cache.TrafficConfigCache.AddTrafficConfigToCache(trafficconfig)
					}()
				}
			}
			wg.Wait()
		})

		AfterEach(func() {
			cache.TrafficConfigCache.Reset()
		})

		It("should be able to get trafficconfigs from cache and pass relevant operations", func() {
			for i := 0; i < noOfTrafficConfigs; i++ {
				identity := fmt.Sprintf("app.exampleasset-%d", i)
				tcEntry := cache.TrafficConfigCache.GetTrafficConfigEntry(identity)
				// Copy the tcEntry to make sure the original tcEntry is not modified and both are equal
				copiedTcEntry := tcEntry.Copy()
				Expect(tcEntry).To(Equal(copiedTcEntry))

				Expect(tcEntry).ToNot(BeNil())
				Expect(len(tcEntry.EnvTrafficConfig)).To(Equal(noOfTrafficConfigEnvs))
				Expect(len(tcEntry.EnvServiceRoutesConfig)).To(Equal(noOfTrafficConfigEnvs))
				trafficconfig.Name = fmt.Sprintf("trafficconfig-%d", i)
				for j := 0; j < noOfTrafficConfigEnvs; j++ {
					trafficconfig.ObjectMeta.Labels["env"] = fmt.Sprintf("env-%d", j)
					trafficconfig.ObjectMeta.Labels["asset"] = "app.exampleasset-" + fmt.Sprintf("%d", i)
					tc := cache.TrafficConfigCache.Get(identity, "env-0")
					Expect(tc).ToNot(BeNil())
				}
			}
			Expect(cache.TrafficConfigCache.GetTotalTrafficConfigs()).To(Equal(noOfTrafficConfigs * noOfTrafficConfigEnvs))
		})
	})

	When("trafficconfigs are deleted from cache", func() {
		BeforeEach(func(sctx SpecContext) {
			for i := 0; i < noOfTrafficConfigs; i++ {
				for j := 0; j < noOfTrafficConfigEnvs; j++ {
					trafficconfig = resourcebuilder.GetFakeTrafficConfig("app.exampleasset-"+fmt.Sprintf("%d", i), fmt.Sprintf("env-%d", j), "1", "namespace")
					cache.TrafficConfigCache.AddTrafficConfigToCache(trafficconfig)
				}
			}
		})

		AfterEach(func() {
			cache.TrafficConfigCache.Reset()
		})

		It("should pass all the operations after tc deletion", func() {
			Expect(cache.TrafficConfigCache.GetTotalTrafficConfigs()).To(Equal(noOfTrafficConfigs * noOfTrafficConfigEnvs))
			for i := 0; i < noOfTrafficConfigs; i++ {
				identity := fmt.Sprintf("app.exampleasset-%d", i)
				for j := 0; j < noOfTrafficConfigEnvs; j++ {
					env := fmt.Sprintf("env-%d", j)
					tc := cache.TrafficConfigCache.Get(identity, env)
					cache.TrafficConfigCache.DeleteTrafficConfigFromCache(tc)
					tc = cache.TrafficConfigCache.Get(identity, env)
					Expect(tc).To(BeNil())
				}
			}
			Expect(cache.TrafficConfigCache.GetTotalTrafficConfigs()).To(Equal(0))
		})
	})
})
