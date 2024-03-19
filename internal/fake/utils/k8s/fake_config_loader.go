package k8s

import (
	"fmt"
	"sync"

	argoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	fakeargoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned/fake"
	admiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"

	k8s_utils "github.com/intuit/naavik/internal/utils/k8s"
	fakeadmiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned/fake"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	fakeistioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func createValidTestConfig(clustername string) *clientcmdapi.Config {
	server := fmt.Sprintf("https://%s.intuit.com:8080", clustername)
	token := "the-token"

	config := clientcmdapi.NewConfig()
	config.Clusters[clustername] = &clientcmdapi.Cluster{
		Server:        server,
		TLSServerName: clustername,
	}
	config.AuthInfos[clustername] = &clientcmdapi.AuthInfo{
		Token: token,
	}
	config.Contexts[clustername] = &clientcmdapi.Context{
		Cluster:  clustername,
		AuthInfo: clustername,
	}
	config.CurrentContext = clustername

	return config
}

var (
	k8sClientMap     sync.Map
	istioClientMap   sync.Map
	argoClientMap    sync.Map
	admiralClientMap sync.Map
)

type FakeClientConfigLoader interface {
	k8s_utils.ClientConfigLoader
	ResetFakeClients()
}

type k8sFakeClientConfigLoader struct{}

func NewFakeConfigLoader() FakeClientConfigLoader {
	return k8sFakeClientConfigLoader{}
}

func (k8sFakeClientConfigLoader) Load(data []byte) (*clientcmdapi.Config, error) {
	return createValidTestConfig(string(data)), nil
}

func (kccl k8sFakeClientConfigLoader) GetConfigFromPath(kubeConfigPath string) (*rest.Config, error) {
	config := createValidTestConfig(string(kubeConfigPath))
	dc := kccl.DefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	return kccl.ClientConfig(dc)
}

func (k8sFakeClientConfigLoader) ClientFromConfig(config *rest.Config) (kubernetes.Interface, error) {
	client, _ := k8sClientMap.LoadOrStore(config.ServerName, k8sfake.NewSimpleClientset())
	return client.(kubernetes.Interface), nil
}

func (k8sFakeClientConfigLoader) RawConfigFromClientConfig(clientConfig clientcmd.ClientConfig) (clientcmdapi.Config, error) {
	return clientConfig.RawConfig()
}

func (kccl k8sFakeClientConfigLoader) ClientFromPath(kubeConfigPath string) (kubernetes.Interface, error) {
	client, _ := k8sClientMap.LoadOrStore(kubeConfigPath, k8sfake.NewSimpleClientset())
	return client.(kubernetes.Interface), nil
}

func (k8sFakeClientConfigLoader) DefaultClientConfig(config clientcmdapi.Config, _ *clientcmd.ConfigOverrides) clientcmd.ClientConfig {
	clientconfig := createValidTestConfig(config.CurrentContext)
	return clientcmd.NewNonInteractiveClientConfig(*clientconfig, config.CurrentContext, &clientcmd.ConfigOverrides{}, nil)
}

func (k8sFakeClientConfigLoader) ClientConfig(clientConfig clientcmd.ClientConfig) (*rest.Config, error) {
	return clientConfig.ClientConfig()
}

func (kccl k8sFakeClientConfigLoader) AdmiralClientFromPath(kubeConfigPath string) (admiralclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.AdmiralClientFromConfig(config)
}

func (kccl k8sFakeClientConfigLoader) AdmiralClientFromConfig(config *rest.Config) (admiralclientset.Interface, error) {
	client, _ := admiralClientMap.LoadOrStore(config.ServerName, fakeadmiralclientset.NewSimpleClientset())
	return client.(admiralclientset.Interface), nil
}

func (kccl k8sFakeClientConfigLoader) IstioClientFromPath(kubeConfigPath string) (istioclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.IstioClientFromConfig(config)
}

func (kccl k8sFakeClientConfigLoader) IstioClientFromConfig(config *rest.Config) (istioclientset.Interface, error) {
	client, _ := istioClientMap.LoadOrStore(config.ServerName, fakeistioclientset.NewSimpleClientset())
	return client.(istioclientset.Interface), nil
}

func (kccl k8sFakeClientConfigLoader) ArgoClientFromPath(kubeConfigPath string) (argoclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.ArgoClientFromConfig(config)
}

func (kccl k8sFakeClientConfigLoader) ArgoClientFromConfig(config *rest.Config) (argoclientset.Interface, error) {
	client, _ := argoClientMap.LoadOrStore(config.ServerName, fakeargoclientset.NewSimpleClientset())
	return client.(argoclientset.Interface), nil
}

func (k8sFakeClientConfigLoader) ResetFakeClients() {
	k8sClientMap = sync.Map{}
	istioClientMap = sync.Map{}
	argoClientMap = sync.Map{}
	admiralClientMap = sync.Map{}
}
