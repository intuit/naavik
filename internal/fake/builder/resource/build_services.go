package resourcebuilder

import (
	"context"
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNFakeServicesInNNamespace(client kubernetes.Interface, serviceNamePrefix string, appName string, namespace string, noOfServices int, noOfNamesapces int) {
	for i := 0; i < noOfNamesapces; i++ {
		namespacePrefix := fmt.Sprintf("%s-%d", namespace, i+1)
		CreateNFakeServices(client, serviceNamePrefix, appName, namespacePrefix, noOfServices)
	}
}

func CreateNFakeServices(client kubernetes.Interface, serviceNamePrefix string, appName string, namespace string, noOfServices int) {
	for i := 0; i < noOfServices; i++ {
		name := fmt.Sprintf("%s-%d", serviceNamePrefix, i+1)
		appName := fmt.Sprintf("%s-%d", appName, i+1)
		CreateFakeService(client, name, appName, namespace)
	}
}

func CreateFakeService(client kubernetes.Interface, name string, appName string, namespace string) {
	assetAlias := fmt.Sprintf("intuit.%s.%s", namespace, name)
	svc := BuildFakeService(name, assetAlias, appName, namespace)
	client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
}

func BuildFakeService(name, assetAlias, appName, namespace string) *corev1.Service {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app":        appName,
			"assetAlias": assetAlias,
		},
		Annotations: map[string]string{
			options.GetWorkloadIdentityKey(): assetAlias,
		},
	}
	svc := &corev1.Service{
		ObjectMeta: objectMeta,
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http-service-mesh",
					Protocol: "TCP",
					Port:     8090,
				},
			},
			Selector: map[string]string{
				"app": appName,
			},
		},
	}
	return svc
}

func BuildFakeHeadlessService(name, assetAlias, appName, namespace string) *corev1.Service {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
		Labels: map[string]string{
			"app":        appName,
			"assetAlias": assetAlias,
		},
		Annotations: map[string]string{
			options.GetWorkloadIdentityKey(): assetAlias,
		},
	}
	svc := &corev1.Service{
		ObjectMeta: objectMeta,
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:     "http-service-mesh",
					Protocol: "TCP",
					Port:     8090,
				},
			},
			Selector: map[string]string{
				"app": appName,
			},
		},
	}
	return svc
}
