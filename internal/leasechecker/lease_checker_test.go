package leasechecker_test

import (
	gocontext "context"
	"testing"

	"github.com/intuit/naavik/internal/leasechecker"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLeaseChecker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "lease_checker_test")
}

var _ = Describe("Test Lease Checker Test", Label("lease_checker_test"), func() {
	BeforeEach(func() {
		leasechecker.ResetState()
	})

	AfterEach(func() {
		leasechecker.ResetState()
	})

	Context("Test NoOp State Checker initialization", func() {
		It("should update the state to read/write by default", func() {
			Expect(leasechecker.IsReadOnly()).To(BeTrue())
			Expect(leasechecker.IsStateInitialized()).To(BeFalse())

			ctx := context.NewContextWithLogger()
			newctx, cancel := gocontext.WithCancel(ctx.Context)
			defer cancel()
			ctx.Context = newctx
			leaseChecker := leasechecker.GetStateChecker(ctx, types.StateCheckerNone)
			leasechecker.RunStateCheck(ctx, leaseChecker)
			leaseChecker.InitStateCache(nil)
			Expect(leasechecker.IsReadOnly()).To(BeFalse())
			Expect(leasechecker.IsStateInitialized()).To(BeTrue())
		})
	})
})
