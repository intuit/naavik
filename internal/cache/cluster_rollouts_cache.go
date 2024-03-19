package cache

import (
	"strings"
	"sync"

	argov1alpha1 "github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/intuit/naavik/pkg/utils"
)

type RolloutCache interface {
	BaseCache
	GetByClusterIdentityEnv(clusterID, identity, env string) *argov1alpha1.Rollout
	GetByClusterIdentity(clusterID, identity string) *RolloutEntry
	GetByIdentity(identity string) []*RolloutItem
	GetNoOfRollouts() int
	Add(clusterID string, rollout *argov1alpha1.Rollout)
	Delete(clusterID string, rollout *argov1alpha1.Rollout)
}

var (
	rolloutcache     = sync.Map{} // map[clusterID]rolloutIdentityMap
	rolloutcacheLock = sync.RWMutex{}
)

type (
	rolloutIdentityMapEntry = map[string]*RolloutEntry // map[identity]*RolloutEntry
)

type RolloutEntry struct {
	Identity string                  `json:"identity"`
	Rollouts map[string]*RolloutItem // map[env]*RolloutItem
}

type RolloutItem struct {
	Rollout   *argov1alpha1.Rollout `json:"rollout"`
	ClusterID string                `json:"clusterId"`
}

var Rollouts = newRolloutsCache()

type rolloutClusterCache struct{}

func newRolloutsCache() RolloutCache {
	return &rolloutClusterCache{}
}

func (rolloutClusterCache) GetByClusterIdentityEnv(clusterID, identity, env string) *argov1alpha1.Rollout {
	clusterID = strings.ToLower(clusterID)
	identity = strings.ToLower(identity)
	env = strings.ToLower(env)
	rolloutcacheLock.RLock()
	defer rolloutcacheLock.RUnlock()
	rolloutIdentityMapEntryVal, ok := rolloutcache.Load(clusterID)
	if !ok {
		return nil
	}
	rolloutIdentityMap := rolloutIdentityMapEntryVal.(rolloutIdentityMapEntry)
	rolloutEntry, ok := rolloutIdentityMap[identity]
	if !ok {
		return nil
	}
	rolloutItem, ok := rolloutEntry.Rollouts[env]
	if !ok {
		return nil
	}
	return rolloutItem.Rollout
}

func (rolloutClusterCache) GetByClusterIdentity(clusterID, identity string) *RolloutEntry {
	clusterID = strings.ToLower(clusterID)
	identity = strings.ToLower(identity)
	rolloutcacheLock.RLock()
	defer rolloutcacheLock.RUnlock()
	rolloutIdentityMapEntryVal, ok := rolloutcache.Load(clusterID)
	if !ok {
		return nil
	}
	rolloutIdentityMap := rolloutIdentityMapEntryVal.(rolloutIdentityMapEntry)
	rolloutEntry, ok := rolloutIdentityMap[identity]
	if !ok {
		return nil
	}
	return rolloutEntry
}

func (rolloutClusterCache) GetByIdentity(identity string) []*RolloutItem {
	identity = strings.ToLower(identity)
	list := []*RolloutItem{}
	rolloutcacheLock.RLock()
	defer rolloutcacheLock.RUnlock()
	rolloutcache.Range(func(key, value interface{}) bool {
		rolloutIdentityMap := value.(rolloutIdentityMapEntry)
		rolloutEntry, ok := rolloutIdentityMap[identity]
		if !ok {
			return true
		}
		for _, rolloutItem := range rolloutEntry.Rollouts {
			list = append(list, rolloutItem)
		}
		return true
	})
	return list
}

func (rolloutClusterCache) Add(clusterID string, rollout *argov1alpha1.Rollout) {
	clusterID = strings.ToLower(clusterID)
	identity := strings.ToLower(utils.ResourceUtil().GetWorkloadIdentifier(rollout.Spec.Template.ObjectMeta))
	env := strings.ToLower(utils.ResourceUtil().GetEnv(rollout.Spec.Template.ObjectMeta, rollout.Name, rollout.Namespace))
	if len(identity) == 0 || len(env) == 0 {
		return
	}
	rolloutcacheLock.Lock()
	defer rolloutcacheLock.Unlock()
	rolloutIdentityMapEntryVal, ok := rolloutcache.Load(clusterID)
	if !ok {
		rolloutIdentityMap := rolloutIdentityMapEntry{}
		rolloutIdentityMap[identity] = &RolloutEntry{
			Identity: identity,
			Rollouts: map[string]*RolloutItem{
				env: {
					Rollout:   rollout,
					ClusterID: clusterID,
				},
			},
		}
		rolloutcache.Store(clusterID, rolloutIdentityMap)
		return
	}
	rolloutIdentityMap := rolloutIdentityMapEntryVal.(rolloutIdentityMapEntry)
	rolloutEntry, ok := rolloutIdentityMap[identity]
	if !ok {
		rolloutIdentityMap[identity] = &RolloutEntry{
			Identity: identity,
			Rollouts: map[string]*RolloutItem{
				env: {
					Rollout:   rollout,
					ClusterID: clusterID,
				},
			},
		}
		return
	}
	rolloutEntry.Rollouts[env] = &RolloutItem{
		Rollout:   rollout,
		ClusterID: clusterID,
	}
}

func (rolloutClusterCache) Delete(clusterID string, rollout *argov1alpha1.Rollout) {
	clusterID = strings.ToLower(clusterID)
	identity := strings.ToLower(utils.ResourceUtil().GetWorkloadIdentifier(rollout.Spec.Template.ObjectMeta))
	env := strings.ToLower(utils.ResourceUtil().GetEnv(rollout.Spec.Template.ObjectMeta, rollout.Name, rollout.Namespace))
	rolloutIdentityMapEntryVal, ok := rolloutcache.Load(clusterID)
	if !ok {
		return
	}
	rolloutcacheLock.Lock()
	defer rolloutcacheLock.Unlock()
	rolloutIdentityMap := rolloutIdentityMapEntryVal.(rolloutIdentityMapEntry)
	rolloutEntry, ok := rolloutIdentityMap[identity]
	if !ok {
		return
	}
	delete(rolloutEntry.Rollouts, env)
	if len(rolloutEntry.Rollouts) == 0 {
		delete(rolloutIdentityMap, identity)
	}
	if len(rolloutIdentityMap) == 0 {
		rolloutcache.Delete(clusterID)
	}
}

func (rolloutClusterCache) GetNoOfRollouts() int {
	count := 0
	rolloutcacheLock.RLock()
	defer rolloutcacheLock.RUnlock()
	rolloutcache.Range(func(key, value interface{}) bool {
		rolloutIdentityMap := value.(rolloutIdentityMapEntry)
		for _, depEntry := range rolloutIdentityMap {
			count += len(depEntry.Rollouts)
		}
		return true
	})
	return count
}

// Reset resets the cache. This is only used for testing.
func (rolloutClusterCache) Reset() {
	rolloutcacheLock.Lock()
	defer rolloutcacheLock.Unlock()
	rolloutcache = sync.Map{}
}
