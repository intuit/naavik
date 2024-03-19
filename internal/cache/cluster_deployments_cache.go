package cache

import (
	"strings"
	"sync"

	"github.com/intuit/naavik/pkg/utils"
	k8sAppsV1 "k8s.io/api/apps/v1"
)

type DeploymentCache interface {
	BaseCache
	GetByClusterIdentityEnv(clusterID, identity, env string) *k8sAppsV1.Deployment
	GetByClusterIdentity(clusterID, identity string) *DeploymentEntry
	GetByIdentity(identity string) []*DeploymentItem
	GetNoOfDeployments() int
	Add(clusterID string, deployment *k8sAppsV1.Deployment)
	Delete(clusterID string, deployment *k8sAppsV1.Deployment)
}

var (
	depcache     = sync.Map{} // map[clusterID]deploymentIdentityMap
	depcacheLock = sync.RWMutex{}
)

type (
	deploymentIdentityMapEntry = map[string]*DeploymentEntry // map[identity]*DeploymentEntry
)

type DeploymentEntry struct {
	Identity    string                     `json:"identity"`
	Deployments map[string]*DeploymentItem // map[env]*DeploymentItem
}

type DeploymentItem struct {
	Deployment *k8sAppsV1.Deployment `json:"deployment"`
	ClusterID  string                `json:"clusterId"`
}

var Deployments = newDeploymentsCache()

type deploymentClusterCache struct{}

func newDeploymentsCache() DeploymentCache {
	return &deploymentClusterCache{}
}

func (deploymentClusterCache) GetByClusterIdentityEnv(clusterID, identity, env string) *k8sAppsV1.Deployment {
	clusterID = strings.ToLower(clusterID)
	identity = strings.ToLower(identity)
	env = strings.ToLower(env)

	depcacheLock.RLock()
	defer depcacheLock.RUnlock()
	depIdentityMapEntry, ok := depcache.Load(clusterID)
	if !ok {
		return nil
	}
	depIdentityMap := depIdentityMapEntry.(deploymentIdentityMapEntry)
	depEntry, ok := depIdentityMap[identity]
	if !ok {
		return nil
	}
	depItem, ok := depEntry.Deployments[env]
	if !ok {
		return nil
	}
	// Create new copy
	deployment := *depItem.Deployment
	return &deployment
}

func (deploymentClusterCache) GetByClusterIdentity(clusterID, identity string) *DeploymentEntry {
	clusterID = strings.ToLower(clusterID)
	identity = strings.ToLower(identity)
	depcacheLock.RLock()
	defer depcacheLock.RUnlock()
	depIdentityMapEntry, ok := depcache.Load(clusterID)
	if !ok {
		return nil
	}
	depIdentityMap := depIdentityMapEntry.(deploymentIdentityMapEntry)
	depEntry, ok := depIdentityMap[identity]
	if !ok {
		return nil
	}
	depEntryCopy := *depEntry
	return &depEntryCopy
}

func (deploymentClusterCache) GetByIdentity(identity string) []*DeploymentItem {
	identity = strings.ToLower(identity)
	list := []*DeploymentItem{}
	depcacheLock.RLock()
	defer depcacheLock.RUnlock()
	depcache.Range(func(key, value interface{}) bool {
		depIdentityMap := value.(deploymentIdentityMapEntry)
		depEntry, ok := depIdentityMap[identity]
		if !ok {
			return true
		}
		for _, depItem := range depEntry.Deployments {
			copyDepItem := *depItem
			list = append(list, &copyDepItem)
		}
		return true
	})
	return list
}

func (deploymentClusterCache) Add(clusterID string, deployment *k8sAppsV1.Deployment) {
	clusterID = strings.ToLower(clusterID)
	identity := strings.ToLower(utils.ResourceUtil().GetWorkloadIdentifier(deployment.Spec.Template.ObjectMeta))
	env := strings.ToLower(utils.ResourceUtil().GetEnv(deployment.Spec.Template.ObjectMeta, deployment.Name, deployment.Namespace))
	if len(identity) == 0 || len(env) == 0 {
		return
	}
	depcacheLock.Lock()
	defer depcacheLock.Unlock()
	depIdentityMapEntry, ok := depcache.Load(clusterID)
	if !ok {
		depIdentityMap := deploymentIdentityMapEntry{}
		depIdentityMap[identity] = &DeploymentEntry{
			Identity: identity,
			Deployments: map[string]*DeploymentItem{
				env: {
					Deployment: deployment,
					ClusterID:  clusterID,
				},
			},
		}
		depcache.Store(clusterID, depIdentityMap)
		return
	}
	depIdentityMap := depIdentityMapEntry.(deploymentIdentityMapEntry)
	depEntry, ok := depIdentityMap[identity]
	if !ok {
		depIdentityMap[identity] = &DeploymentEntry{
			Identity: identity,
			Deployments: map[string]*DeploymentItem{
				env: {
					Deployment: deployment,
					ClusterID:  clusterID,
				},
			},
		}
		return
	}
	depEntry.Deployments[env] = &DeploymentItem{
		Deployment: deployment,
		ClusterID:  clusterID,
	}
}

func (deploymentClusterCache) Delete(clusterID string, deployment *k8sAppsV1.Deployment) {
	clusterID = strings.ToLower(clusterID)
	identity := strings.ToLower(utils.ResourceUtil().GetWorkloadIdentifier(deployment.Spec.Template.ObjectMeta))
	env := strings.ToLower(utils.ResourceUtil().GetEnv(deployment.Spec.Template.ObjectMeta, deployment.Name, deployment.Namespace))
	depcacheLock.Lock()
	defer depcacheLock.Unlock()
	depIdentityMapEntry, ok := depcache.Load(clusterID)
	if !ok {
		return
	}
	depIdentityMap := depIdentityMapEntry.(deploymentIdentityMapEntry)
	depEntry, ok := depIdentityMap[identity]
	if !ok {
		return
	}
	delete(depEntry.Deployments, env)
	if len(depEntry.Deployments) == 0 {
		delete(depIdentityMap, identity)
	}
	if len(depIdentityMap) == 0 {
		depcache.Delete(clusterID)
	}
}

func (deploymentClusterCache) GetNoOfDeployments() int {
	count := 0
	depcacheLock.RLock()
	defer depcacheLock.RUnlock()
	depcache.Range(func(key, value interface{}) bool {
		depIdentityMap := value.(deploymentIdentityMapEntry)
		for _, depEntry := range depIdentityMap {
			count += len(depEntry.Deployments)
		}
		return true
	})
	return count
}

// Reset resets the cache. This is only used for testing.
func (deploymentClusterCache) Reset() {
	depcacheLock.Lock()
	defer depcacheLock.Unlock()
	depcache = sync.Map{}
}
