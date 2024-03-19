package resourcebuilder

import (
	"context"
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	admiralApi "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	clientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateNFakeDependencies(
	client clientset.Interface,
	noOfClusters int,
	noOfNamespaces int,
	namespacePrefix string,
	workloadName string,
	maxDependencyNetwork int,
) {
	// There can only be one deploy/rollout per namespace
	name := workloadName
	for i := 1; i < noOfClusters+1; i++ {
		cluster := fmt.Sprintf("cluster-%d", i)
		for j := 1; j < noOfNamespaces+1; j++ {
			namespace := fmt.Sprintf("%s-%d", namespacePrefix, j)
			source := fmt.Sprintf("app.%s.%s.%s", cluster, namespace, name)
			destinations := []string{}
			for k := 1; k < maxDependencyNetwork+1; k++ {
				destinationCluster := fmt.Sprintf("cluster-%d", i+1)
				namespace := fmt.Sprintf("%s-%d", namespacePrefix, j+k)
				destination := fmt.Sprintf("app.%s.%s.%s", destinationCluster, namespace, name)
				destinations = append(destinations, destination)
			}
			CreateFakeDependency(client, options.GetDependenciesNamespace(), source, destinations)
		}
	}
}

func CreateFakeDependency(client clientset.Interface, namespace string, source string, destinations []string) {
	dep := BuildFakeDependency(namespace, source, destinations)
	client.AdmiralV1().Dependencies(namespace).Create(context.Background(), dep, metav1.CreateOptions{})
}

func BuildFakeDependency(namespace string, source string, destinations []string) *admiralApi.Dependency {
	return &admiralApi.Dependency{
		ObjectMeta: metav1.ObjectMeta{
			Name:      source,
			Namespace: namespace,
		},
		Spec: admiralApi.DependencySpec{
			Source:        source,
			IdentityLabel: source,
			Destinations:  destinations,
		},
	}
}
