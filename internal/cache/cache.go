package cache

import fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"

type BaseCache interface {
	// This is used for testing purposes only.
	// Reset resets the cache.
	Reset()
}

// This is used for testing purposes only.
// ResetAllCache resets all caches.
func ResetAllCaches() {
	Deployments.Reset()
	Rollouts.Reset()
	Services.Reset()
	ControllerCache.Reset()
	IdentityCluster.Reset()
	IdentityDependency.Reset()
	RemoteCluster.Reset()
	TrafficConfigCache.Reset()
	fake_k8s_utils.NewFakeConfigLoader().ResetFakeClients()
}
