package k8sutils

import (
	"fmt"

	argoclientset "github.com/argoproj/argo-rollouts/pkg/client/clientset/versioned"
	admiralclientset "github.com/istio-ecosystem/admiral-api/pkg/client/clientset/versioned"
	istioclientset "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClientConfigLoader interface {
	Load(data []byte) (*clientcmdapi.Config, error)
	ClientFromConfig(config *rest.Config) (kubernetes.Interface, error)
	ClientFromPath(kubeConfigPath string) (kubernetes.Interface, error)
	GetConfigFromPath(kubeConfigPath string) (*rest.Config, error)
	DefaultClientConfig(config clientcmdapi.Config, overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig
	RawConfigFromClientConfig(clientConfig clientcmd.ClientConfig) (clientcmdapi.Config, error)
	ClientConfig(clientConfig clientcmd.ClientConfig) (*rest.Config, error)

	AdmiralClientFromConfig(config *rest.Config) (admiralclientset.Interface, error)
	AdmiralClientFromPath(kubeConfigPath string) (admiralclientset.Interface, error)

	IstioClientFromPath(kubeConfigPath string) (istioclientset.Interface, error)
	IstioClientFromConfig(config *rest.Config) (istioclientset.Interface, error)

	ArgoClientFromPath(kubeConfigPath string) (argoclientset.Interface, error)
	ArgoClientFromConfig(config *rest.Config) (argoclientset.Interface, error)
}

type k8sClientConfigLoader struct{}

func NewConfigLoader() ClientConfigLoader {
	return k8sClientConfigLoader{}
}

func (k8sClientConfigLoader) Load(data []byte) (*clientcmdapi.Config, error) {
	return clientcmd.Load(data)
}

func (k8sClientConfigLoader) GetConfigFromPath(kubeConfigPath string) (*rest.Config, error) {
	// create the config from the path
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil || config == nil {
		return nil, fmt.Errorf("could not retrieve kubeconfig: %v", err)
	}
	return config, err
}

func (k8sClientConfigLoader) ClientFromConfig(config *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(config)
}

func (k8sClientConfigLoader) RawConfigFromClientConfig(clientConfig clientcmd.ClientConfig) (clientcmdapi.Config, error) {
	return clientConfig.RawConfig()
}

func (kccl k8sClientConfigLoader) ClientFromPath(kubeConfigPath string) (kubernetes.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.ClientFromConfig(config)
}

func (k8sClientConfigLoader) DefaultClientConfig(config clientcmdapi.Config, overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(config, overrides)
}

func (k8sClientConfigLoader) ClientConfig(clientConfig clientcmd.ClientConfig) (*rest.Config, error) {
	return clientConfig.ClientConfig()
}

func (kccl k8sClientConfigLoader) AdmiralClientFromPath(kubeConfigPath string) (admiralclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.AdmiralClientFromConfig(config)
}

func (kccl k8sClientConfigLoader) AdmiralClientFromConfig(config *rest.Config) (admiralclientset.Interface, error) {
	return admiralclientset.NewForConfig(config)
}

func (kccl k8sClientConfigLoader) IstioClientFromConfig(config *rest.Config) (istioclientset.Interface, error) {
	return istioclientset.NewForConfig(config)
}

func (kccl k8sClientConfigLoader) IstioClientFromPath(kubeConfigPath string) (istioclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.IstioClientFromConfig(config)
}

func (kccl k8sClientConfigLoader) ArgoClientFromConfig(config *rest.Config) (argoclientset.Interface, error) {
	return argoclientset.NewForConfig(config)
}

func (kccl k8sClientConfigLoader) ArgoClientFromPath(kubeConfigPath string) (argoclientset.Interface, error) {
	config, err := kccl.GetConfigFromPath(kubeConfigPath)
	if err != nil || config == nil {
		return nil, err
	}
	return kccl.ArgoClientFromConfig(config)
}
