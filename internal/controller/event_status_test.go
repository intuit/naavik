package controller_test

import (
	"errors"

	"github.com/intuit/naavik/internal/controller"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test event status", Label("event_status_test"), func() {
	When("New event status is created", func() {
		It("should create new event status", func() {
			eps := controller.NewEventProcessStatus()
			Expect(eps.Status).To(Equal(controller.EventCompleted))
			Expect(eps.MaxRetryCount).To(Equal(controller.DefaultMaxRetryCount))
		})

		It("should create new event status with status", func() {
			eps := controller.NewEventProcessStatus().WithStatus(controller.EventProcessing)
			Expect(eps.Status).To(Equal(controller.EventProcessing))
		})

		It("should create new event status with retry", func() {
			eps := controller.NewEventProcessStatus().WithRetry()
			Expect(eps.Retry).To(BeTrue())
		})

		It("should create new event status with max retry", func() {
			eps := controller.NewEventProcessStatus().WithMaxRetry(10)
			Expect(eps.MaxRetryCount).To(Equal(10))
		})

		It("should create new event status with message", func() {
			eps := controller.NewEventProcessStatus().WithMessage("test", "test")
			Expect(eps.Message["test"]).To(Equal("test"))
		})

		It("should create new event status with error", func() {
			eps := controller.NewEventProcessStatus().WithError(errors.New("test error"))
			Expect(eps.Error.Error()).To(Equal("test error"))
		})

		It("should create new event status with nil channel", func() {
			eps := controller.NewEventProcessStatus()
			eps.SkipClose(nil)
		})

		It("should create new event status with channel", func() {
			eps := controller.NewEventProcessStatus().WithStatus(controller.EventFailure).WithMessage("test", "val")
			ch := make(chan controller.EventProcessStatus, 1)
			eps.Send(ch)
			receivedEps := <-ch
			Eventually(receivedEps).Should(Equal(eps))

			eps.WithStatus(controller.EventFailure).WithMessage("test", "val").SkipClose(ch)
			receivedEps = <-ch
			eps.Status = controller.EventSkip
			Eventually(receivedEps).Should(Equal(eps))
			Expect(func() { eps.WithStatus(controller.EventFailure).WithMessage("test", "val").Send(ch) }).To(Panic())
		})
	})
})
