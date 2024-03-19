package perf

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNaavikPerf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "traffic_config_perf")
}
