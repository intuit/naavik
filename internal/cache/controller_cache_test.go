package cache_test

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Controllers Cache", Label("controllers_cache_test"), func() {
	BeforeEach(func(sctx SpecContext) {
		options.InitializeNaavikArgs(nil)
	})

	When("controllers are registered to cache", func() {
		BeforeEach(func() {
			cache.ControllerCache.Reset()
		})

		AfterEach(func() {
			cache.ControllerCache.Reset()
		})

		It("should have all registered controllers in cache", func() {
			controllers := []string{"controller1", "conTroller2", "controller2", "controller3"}
			for _, controller := range controllers {
				cache.ControllerCache.Register(controller, cache.ControllerContext{})
			}

			Expect(cache.ControllerCache.List()).To(HaveLen(3))
			Expect(cache.ControllerCache.GetStopCh("controller1")).ToNot(BeNil())
			Expect(cache.ControllerCache.GetStopCh("controller2")).ToNot(BeNil())
			Expect(cache.ControllerCache.GetStopCh("controller3")).ToNot(BeNil())
			Expect(cache.ControllerCache.GetStopCh("controller4")).To(Equal(cache.ControllerContext{}))
			cache.ControllerCache.Range(func(controller any, context interface{}) bool {
				Expect(controllers).To(ContainElement(controller))
				return true
			})
		})
	})

	When("controllers are deregistered to cache", func() {
		BeforeEach(func() {
			cache.ControllerCache.Reset()
		})

		AfterEach(func() {
			cache.ControllerCache.Reset()
		})

		It("should have all registered controllers in cache", func() {
			controllers := []string{"controller1", "conTroller2", "controller2", "controller3"}
			for _, controller := range controllers {
				cache.ControllerCache.Register(controller, cache.ControllerContext{})
			}
			Expect(cache.ControllerCache.List()).To(HaveLen(3))
			cache.ControllerCache.DeRegister("controller1")
			Expect(cache.ControllerCache.List()).To(HaveLen(2))
			cache.ControllerCache.DeRegister("conTroller2")
			Expect(cache.ControllerCache.List()).To(HaveLen(1))
			cache.ControllerCache.DeRegister("controller3")
			Expect(cache.ControllerCache.List()).To(HaveLen(0))
		})
	})
})
