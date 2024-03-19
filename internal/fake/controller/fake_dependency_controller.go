package controller

import (
	"time"

	admiral_controller "github.com/intuit/naavik/internal/controller/admiral"
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	k8s_handlers "github.com/intuit/naavik/internal/handler/dependency"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func NewFakeDependencyController(config *rest.Config, dependencyNamespace string, resync time.Duration) {
	admiral_controller.NewDependencyController(
		config.ServerName,
		config,
		fake_k8s_utils.NewFakeConfigLoader(),
		dependencyNamespace,
		metav1.ListOptions{},
		resync,
		k8s_handlers.NewDependencyHandler(k8s_handlers.Opts{}),
	)
}
