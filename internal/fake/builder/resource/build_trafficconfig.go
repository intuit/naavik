package resourcebuilder

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/pkg/logger"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	admiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:embed fake_tc_data.json
var content embed.FS

func CreateFakeTrafficConfigsInNNamespaceNClusters(
	client admiralclientset.Interface,
	trafficConfigNamespace string,
	noOfClusters int,
	noOfNamespaces int,
	clusterPrefix string,
	namespacePrefix string,
	workloadName string,
	env string,
	revision string,
) map[string]time.Time {
	creationTimestamps := map[string]time.Time{}
	for i := 0; i < noOfClusters; i++ {
		cluster := fmt.Sprintf("%s-%d", clusterPrefix, i+1)
		for j := 0; j < noOfNamespaces; j++ {
			namespace := fmt.Sprintf("%s-%d", namespacePrefix, j+1)
			assetAlias := fmt.Sprintf("app.%s.%s.%s", cluster, namespace, workloadName)
			creationTimestamps[assetAlias] = time.Now()
			CreateFakeTrafficConfig(client, assetAlias, env, revision, trafficConfigNamespace)
		}
	}
	return creationTimestamps
}

func CreateFakeTrafficConfig(client admiralclientset.Interface, assetAlias string, env string, revision string, trafficConfigNamespace string) {
	tc := GetPreloadedTrafficConfig()
	tc.ObjectMeta.Annotations["revisionNumber"] = revision
	tc.ObjectMeta.Labels["asset"] = assetAlias
	tc.ObjectMeta.Labels["env"] = env
	tc.Name = fmt.Sprintf("%s-%s", assetAlias, env)
	tc.Namespace = trafficConfigNamespace
	tc.Spec.WorkloadEnv = []string{env}
	tc.Spec.EdgeService.Routes[0].WorkloadEnvSelectors = []string{env}
	tc.Spec.EdgeService.Routes[1].WorkloadEnvSelectors = []string{env}
	tc.Spec.QuotaGroup.TotalQuotaGroup[0].WorkloadEnvSelectors = []string{env}
	tc.Spec.EdgeService.Targets[0].MeshDNS = fmt.Sprintf("%s.%s.%s", env, assetAlias, options.GetHostnameSuffix())
	exists, _ := client.AdmiralV1().TrafficConfigs(trafficConfigNamespace).Get(context.Background(), tc.Name, metav1.GetOptions{})
	if exists != nil {
		tc.ResourceVersion = exists.ResourceVersion
		tc.Generation = exists.Generation + 1
		client.AdmiralV1().TrafficConfigs(trafficConfigNamespace).Update(context.Background(), tc, metav1.UpdateOptions{})
	} else {
		client.AdmiralV1().TrafficConfigs(trafficConfigNamespace).Create(context.Background(), tc, metav1.CreateOptions{})
	}
}

func GetFakeTrafficConfig(assetAlias string, env string, revision string, trafficConfigNamespace string) *admiralv1.TrafficConfig {
	tc := GetPreloadedTrafficConfig()
	tc.ObjectMeta.Annotations["revisionNumber"] = revision
	tc.ObjectMeta.Labels[options.GetTrafficConfigIdentityKey()] = assetAlias
	tc.ObjectMeta.Labels["env"] = env
	tc.Name = fmt.Sprintf("%s-%s", assetAlias, env)
	tc.Namespace = trafficConfigNamespace
	tc.Spec.WorkloadEnv = []string{env}
	tc.Spec.EdgeService.Routes[0].WorkloadEnvSelectors = []string{env}
	tc.Spec.EdgeService.Routes[1].WorkloadEnvSelectors = []string{env}
	tc.Spec.QuotaGroup.TotalQuotaGroup[0].WorkloadEnvSelectors = []string{env}
	tc.Spec.EdgeService.Targets[0].MeshDNS = fmt.Sprintf("%s.%s.%s", env, assetAlias, options.GetHostnameSuffix())
	return tc
}

func GetPreloadedTrafficConfig() *admiralv1.TrafficConfig {
	tc := &admiralv1.TrafficConfig{}
	tcData, err := content.ReadFile("fake_tc_data.json")
	if err != nil {
		logger.Log.Infof("Error reading fake_tc_data.json: %v", err)
	}
	err = json.Unmarshal(tcData, tc)
	if err != nil {
		panic(err)
	}
	return tc
}
