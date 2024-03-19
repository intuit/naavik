package builder

import (
	fake_k8s_utils "github.com/intuit/naavik/internal/fake/utils/k8s"
	"github.com/intuit/naavik/internal/types/remotecluster"
	"k8s.io/client-go/tools/clientcmd"
)

func BuildRemoteCluster(name string) remotecluster.RemoteCluster {
	k8sConfig, _ := fake_k8s_utils.NewFakeConfigLoader().Load([]byte(name))
	clientCmd := fake_k8s_utils.NewFakeConfigLoader().DefaultClientConfig(*k8sConfig, &clientcmd.ConfigOverrides{})
	config, _ := fake_k8s_utils.NewFakeConfigLoader().GetConfigFromPath(name)
	k8sclient, _ := fake_k8s_utils.NewFakeConfigLoader().ClientFromConfig(config)
	istioclient, _ := fake_k8s_utils.NewFakeConfigLoader().IstioClientFromConfig(config)
	admiralClient, _ := fake_k8s_utils.NewFakeConfigLoader().AdmiralClientFromConfig(config)
	argoClient, _ := fake_k8s_utils.NewFakeConfigLoader().ArgoClientFromConfig(config)

	return remotecluster.CreateRemoteCluster(name, name, name, name, clientCmd, k8sclient, istioclient, argoClient, admiralClient)
}
