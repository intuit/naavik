package controller

import (
	"time"

	k8s_controller "github.com/intuit/naavik/internal/controller/k8s"
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	"github.com/intuit/naavik/internal/handler/remotecluster"
	"github.com/intuit/naavik/internal/handler/remotecluster/resolver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func NewFakeSecretController(config *rest.Config, secretNamespace string, resync time.Duration) {
	k8s_controller.NewSecretController(
		config.ServerName,
		config,
		fake_k8s_utils.NewFakeConfigLoader(),
		secretNamespace,
		metav1.ListOptions{},
		resync,
		remotecluster.NewRemoteClusterSecretHandler(remotecluster.SecretHandlerOpts{
			RemoteClusterResolver: remotecluster.NewRemoteClusterResolver(resolver.NewSecretConfigResolver(), fake_k8s_utils.NewFakeConfigLoader()),
		}),
	)
}
