package resourcebuilder

import (
	"context"
	"fmt"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	argoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"github.com/intuit/naavik/cmd/options"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateNFakeRolloutsInNNamespace(client argoclientset.Interface, rolloutNamePrefix string,
	appNamePrefix string, namespace string, cluster string, env string, noOfrollouts int, noOfNamespace int,
) {
	for i := 0; i < noOfNamespace; i++ {
		namespacePrefix := fmt.Sprintf("%s-%d", namespace, i+1)
		CreateNFakeRollouts(client, rolloutNamePrefix, appNamePrefix, namespacePrefix, cluster, env, noOfrollouts)
	}
}

// Creates fake rollouts in N number of namespaces, one deployment per namespace.
func CreateNFakeRollouts(client argoclientset.Interface, rolloutNamePrefix string, appNamePrefix string, namespace string, cluster string, env string, noOfrollouts int) {
	for i := 0; i < noOfrollouts; i++ {
		name := fmt.Sprintf("%s-%d", rolloutNamePrefix, i+1)
		appNamePrefix := fmt.Sprintf("%s-%d", appNamePrefix, i+1)
		CreateFakeRollout(client, name, appNamePrefix, namespace, cluster, env)
	}
}

func CreateFakeRollout(client argoclientset.Interface, name string, app string, namespace string, cluster string, env string) {
	assetAlias := fmt.Sprintf("intuit.%s.%s.%s", cluster, namespace, name)
	rollout := BuildFakeRollout(name, assetAlias, app, env, namespace)
	client.ArgoprojV1alpha1().Rollouts(namespace).Create(context.Background(), rollout, metav1.CreateOptions{})
}

func BuildFakeRollout(name, assetAlias, appName, env, namespace string) *argov1alpha1.Rollout {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app":            appName,
			"assetAlias":     assetAlias,
			"env":            env,
			"istio-injected": "true",
			"express":        "true",
		},
		Annotations: map[string]string{
			options.GetEnvKey():                            env,
			options.GetWorkloadIdentityKey():               assetAlias,
			options.GetMeshInjectionEnabledKey():           "true",
			"traffic.sidecar.istio.io/includeInboundPorts": "8090",
		},
	}
	rollout := &argov1alpha1.Rollout{
		ObjectMeta: objectMeta,
		Spec: argov1alpha1.RolloutSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: objectMeta,
			},
		},
	}
	return rollout
}
