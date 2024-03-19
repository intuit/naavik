package trafficconfig

import (
	"fmt"
	"strings"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/server/api"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	remotecluster "github.com/intuit/naavik/internal/types/remotecluster"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resources struct {
	ClusterResources map[string]map[string]interface{} `json:"clusterResources"` // map[clusterName]map[resourceName]resource
}

func (tcr *Resources) updateVirtualServices(rc remotecluster.RemoteCluster, identity string, env string) (*v1alpha3.VirtualServiceList, error) {
	labelSelector := fmt.Sprintf("%s=%s,%s=%s",
		types.CreatedByKey, types.NaavikName,
		types.CreatedForKey, strings.ToLower(identity))

	if len(env) > 0 {
		labelSelector = fmt.Sprintf("%s=%s,%s=%s,%s=%s",
			types.CreatedByKey, types.NaavikName,
			types.CreatedForKey, strings.ToLower(identity),
			types.CreatedForEnvKey, env)
	}
	// Get VirtualServices from dependent clusters
	virtualservices, err := rc.IstioClient().ListVirtualServices(context.NewContextWithLogger(), options.GetSyncNamespace(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err == nil {
		_, ok := tcr.ClusterResources[rc.GetClusterID()]
		if !ok {
			tcr.ClusterResources[rc.GetClusterID()] = make(map[string]interface{})
		}
		tcr.ClusterResources[rc.GetClusterID()][fmt.Sprintf("%s/%s", identity, "virtualservices")] = virtualservices
	} else {
		tcr.ClusterResources[rc.GetClusterID()][fmt.Sprintf("%s/%s", identity, "virtualservices")] = api.ErrorResponse{Message: err.Error()}
	}
	return virtualservices, err
}

func (tcr *Resources) updateThrottleEnvoyFilters(rc remotecluster.RemoteCluster, identity string, env string) (*v1alpha3.EnvoyFilterList, error) {
	labelSelector := fmt.Sprintf("%s=%s,%s=%s",
		types.CreatedByKey, types.NaavikName,
		types.CreatedForKey, strings.ToLower(identity))

	if len(env) > 0 {
		labelSelector = fmt.Sprintf("%s=%s,%s=%s,%s=%s",
			types.CreatedByKey, types.NaavikName,
			types.CreatedForKey, strings.ToLower(identity),
			types.CreatedForEnvKey, env)
	}
	envoyfilters, err := rc.IstioClient().ListEnvoyFilters(context.NewContextWithLogger(), types.NamespaceIstioSystem, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err == nil {
		_, ok := tcr.ClusterResources[rc.GetClusterID()]
		if !ok {
			tcr.ClusterResources[rc.GetClusterID()] = make(map[string]interface{})
		}
		tcr.ClusterResources[rc.GetClusterID()][fmt.Sprintf("%s/%s", identity, "throttlefilters")] = envoyfilters
	} else {
		tcr.ClusterResources[rc.GetClusterID()][fmt.Sprintf("%s/%s", identity, "throttlefilters")] = api.ErrorResponse{Message: err.Error()}
	}
	return envoyfilters, err
}
