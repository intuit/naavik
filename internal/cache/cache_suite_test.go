package cache_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNaavikCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cache_test")
}
