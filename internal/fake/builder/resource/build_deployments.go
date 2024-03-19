package resourcebuilder

import (
	"context"
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	k8sAppsV1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNFakeDeploymentsInNNamespace(client kubernetes.Interface, deploymentNamePrefix string,
	appNamePrefix string, namespace string, cluster string, env string, noOfDeployments int, noOfNamespace int,
) {
	for i := 0; i < noOfNamespace; i++ {
		namespacePrefix := fmt.Sprintf("%s-%d", namespace, i+1)
		CreateNFakeDeployments(client, deploymentNamePrefix, appNamePrefix, namespacePrefix, cluster, env, noOfDeployments)
	}
}

func CreateNFakeDeployments(client kubernetes.Interface, deploymentNamePrefix string, appNamePrefix string, namespace string, cluster string, env string, noOfDeployments int) {
	for i := 0; i < noOfDeployments; i++ {
		name := fmt.Sprintf("%s-%d", deploymentNamePrefix, i+1)
		appNamePrefix := fmt.Sprintf("%s-%d", appNamePrefix, i+1)
		CreateFakeDeployment(client, name, appNamePrefix, namespace, cluster, env)
	}
}

func CreateFakeDeployment(client kubernetes.Interface, name string, appNamePrefix, namespace string, cluster string, env string) {
	assetAlias := fmt.Sprintf("app.%s.%s.%s", cluster, namespace, name)
	dep := BuildFakeDeployment(name, assetAlias, appNamePrefix, env, namespace)
	client.AppsV1().Deployments(namespace).Create(context.Background(), dep, metav1.CreateOptions{})
}

func BuildFakeDeployment(name, assetAlias, appName, env, namespace string) *k8sAppsV1.Deployment {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app":            appName,
			"assetAlias":     assetAlias,
			"env":            env,
			"istio-injected": "true",
		},
		Annotations: map[string]string{
			options.GetEnvKey():                            env,
			options.GetWorkloadIdentityKey():               assetAlias,
			options.GetMeshInjectionEnabledKey():           "true",
			"traffic.sidecar.istio.io/includeInboundPorts": "8090",
		},
	}
	dep := &k8sAppsV1.Deployment{
		ObjectMeta: objectMeta,
		Spec: k8sAppsV1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: objectMeta,
			},
		},
	}
	return dep
}
