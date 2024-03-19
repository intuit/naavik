package resourcebuilder

import (
	"context"
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateNRemoteClusterSecret(client kubernetes.Interface, n int, clusterNamePrefix string, namespace string) {
	for i := 0; i < n; i++ {
		clusterName := fmt.Sprintf("%s-%d", clusterNamePrefix, i+1)
		logger.Log.Infof("Creating secret %s in namespace %s", clusterName, namespace)
		secret := CreateRemoteClusterSecret(clusterName, namespace)
		client.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	}
}

func CreateRemoteClusterSecret(clusterName string, namespace string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: namespace,
			Labels: map[string]string{
				options.GetSecretSyncLabel(): "true",
			},
		},
		Data: map[string][]byte{
			clusterName: []byte(clusterName),
		},
	}
	return secret
}
