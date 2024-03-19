package cache

import (
	"strings"
	"sync"

	"github.com/intuit/naavik/pkg/types/set"
)

type IdentityClusterCache interface {
	BaseCache
	AddClusterToIdentity(identity string, clusterID string)
	DeleteClusterFromIdentity(identity string, clusterID string)
	GetClustersForIdentity(identity string) []string
	ListIdentities() []string
	IsClusterPresentInIdentity(identity string, clusterID string) bool
}

var IdentityCluster = newIdentityCluster()

type identityCluster struct{}

var identityClusterLock = sync.RWMutex{}

var identityClusterCache = sync.Map{} // Map[clusterID]IdentityCluster

func newIdentityCluster() IdentityClusterCache {
	return &identityCluster{}
}

func (*identityCluster) AddClusterToIdentity(identity string, clusterID string) {
	identity = strings.ToLower(identity)
	clusterID = strings.ToLower(clusterID)
	identityClusterLock.Lock()
	defer identityClusterLock.Unlock()
	existing, loaded := identityClusterCache.Load(identity)
	if loaded {
		existing.(set.Set[string]).Add(clusterID)
		identityClusterCache.Store(identity, existing)
	} else {
		clusterIds := set.NewSet[string]()
		clusterIds.Add(clusterID)
		identityClusterCache.Store(identity, clusterIds)
	}
}

func (*identityCluster) DeleteClusterFromIdentity(identity string, clusterID string) {
	identity = strings.ToLower(identity)
	clusterID = strings.ToLower(clusterID)
	identityClusterLock.Lock()
	defer identityClusterLock.Unlock()
	existing, found := identityClusterCache.Load(identity)
	if !found {
		return
	}
	clusterIds := existing.(set.Set[string])
	clusterIds.Delete(clusterID)
}

func (identityCluster) GetClustersForIdentity(identity string) []string {
	identity = strings.ToLower(identity)
	identityClusterLock.RLock()
	defer identityClusterLock.RUnlock()
	existing, found := identityClusterCache.Load(identity)
	if !found {
		return []string{}
	}
	existingClusters := existing.(set.Set[string]).Items()
	clone := make([]string, len(existingClusters))
	copy(clone, existingClusters)
	return clone
}

func (identityCluster) ListIdentities() []string {
	list := []string{}
	identityClusterCache.Range(func(key, value interface{}) bool {
		list = append(list, key.(string))
		return true
	})
	return list
}

func (ic identityCluster) IsClusterPresentInIdentity(identity string, clusterID string) bool {
	identity = strings.ToLower(identity)
	clusterID = strings.ToLower(clusterID)
	identityClusterLock.RLock()
	defer identityClusterLock.RUnlock()
	existing, found := identityClusterCache.Load(identity)
	if !found {
		return false
	}
	return existing.(set.Set[string]).Has(clusterID)
}

func (identityCluster) Reset() {
	identityClusterLock.Lock()
	defer identityClusterLock.Unlock()
	identityClusterCache = sync.Map{}
}
