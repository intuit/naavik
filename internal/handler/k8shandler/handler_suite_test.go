package k8shandler

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8sHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "k8s_handler_test")
}
