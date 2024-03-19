package remotecluster

import (
	"strings"

	argoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	"github.com/intuit/naavik/pkg/clients/clientset/istio"
	admiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	"istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type RemoteCluster interface {
	GetClusterID() string
	GetContextName() string
	GetSecretIdentifier() string
	GetHost() string
	GetClientConfig() clientcmd.ClientConfig
	K8sClient() kubernetes.Interface
	IstioClient() istio.ClientInterface
	ArgoClient() argoclientset.Interface
	AdmiralClient() admiralclientset.Interface
}

type remoteCluster struct {
	clusterID        string
	contextName      string
	secretIdentifier string
	host             string
	clientConfig     clientcmd.ClientConfig
	k8sClient        kubernetes.Interface
	istioClient      istio.ClientInterface
	argoClient       argoclientset.Interface
	admiralClient    admiralclientset.Interface
}

func CreateRemoteCluster(clusterID string, contextName string, secretIdentifier string,
	clusterHost string, clientConfig clientcmd.ClientConfig, k8sClient kubernetes.Interface,
	istioclient versioned.Interface, argoClient argoclientset.Interface,
	admiralClient admiralclientset.Interface,
) RemoteCluster {
	return &remoteCluster{
		clusterID:        clusterID,
		contextName:      contextName,
		secretIdentifier: secretIdentifier,
		host:             clusterHost,
		clientConfig:     clientConfig,
		k8sClient:        k8sClient,
		istioClient:      istio.NewIstioClient(clusterID, istioclient),
		argoClient:       argoClient,
		admiralClient:    admiralClient,
	}
}

func (rc *remoteCluster) GetClusterID() string {
	return strings.ToLower(rc.clusterID)
}

func (rc *remoteCluster) GetContextName() string {
	return rc.contextName
}

func (rc *remoteCluster) GetSecretIdentifier() string {
	return rc.secretIdentifier
}

func (rc *remoteCluster) GetHost() string {
	return rc.host
}

func (rc *remoteCluster) GetClientConfig() clientcmd.ClientConfig {
	return rc.clientConfig
}

func (rc *remoteCluster) K8sClient() kubernetes.Interface {
	return rc.k8sClient
}

func (rc *remoteCluster) IstioClient() istio.ClientInterface {
	return rc.istioClient
}

func (rc *remoteCluster) ArgoClient() argoclientset.Interface {
	return rc.argoClient
}

func (rc *remoteCluster) AdmiralClient() admiralclientset.Interface {
	return rc.admiralClient
}
