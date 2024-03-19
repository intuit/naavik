package trafficconfig

import admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"

type ServiceRouteConfig struct {
	WorkloadEnvRevision map[string]string  `json:"workloadEnvRevision,omitempty"`
	ServiceAssetAlias   string             `json:"serviceAssetAlias,omitempty"`
	Routes              []*admiralv1.Route `json:"routes,omitempty"`
}

func (src *ServiceRouteConfig) Copy() *ServiceRouteConfig {
	copySrc := &ServiceRouteConfig{
		ServiceAssetAlias:   src.ServiceAssetAlias,
		WorkloadEnvRevision: make(map[string]string, len(src.WorkloadEnvRevision)),
	}
	for env, rev := range src.WorkloadEnvRevision {
		copySrc.WorkloadEnvRevision[env] = rev
	}
	for _, route := range src.Routes {
		copySrc.Routes = append(copySrc.Routes, route.DeepCopy())
	}
	return copySrc
}

type RouteConfig struct {
	ServicesConfig []*ServiceRouteConfig `json:"servicesRouteConfig"`
}
