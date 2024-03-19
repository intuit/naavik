package controller

import (
	"time"

	admiral_controller "github.com/intuit/naavik/internal/controller/admiral"
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	traffic_config "github.com/intuit/naavik/internal/handler/trafficconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func NewFakeTrafficConfigController(config *rest.Config, dependencyNamespace string, resync time.Duration) {
	admiral_controller.NewTrafficConfigController(
		config.ServerName,
		config,
		fake_k8s_utils.NewFakeConfigLoader(),
		dependencyNamespace,
		metav1.ListOptions{},
		resync,
		traffic_config.NewTrafficConfigHandler(),
	)
}
