package cache

import (
	"strings"
	"sync"

	"github.com/intuit/naavik/internal/types/remotecluster"
)

type RemoteClusterCache interface {
	BaseCache
	AddCluster(remoteCluster remotecluster.RemoteCluster)
	DeleteCluster(clusterID string)
	// GetCluster returns the remote cluster and a boolean indicating if the cluster exists
	GetCluster(clusterID string) (remotecluster.RemoteCluster, bool)
	ListClusters() []remotecluster.RemoteCluster
	UpdateCluster(clusterID string, remoteCluster remotecluster.RemoteCluster) (remotecluster.RemoteCluster, bool)
}

var RemoteCluster = newRemoteCluster()

type remoteClusterCache struct{}

var remoteClusterListCache = sync.Map{} // Map[clusterID]RemoteCluster

func newRemoteCluster() RemoteClusterCache {
	return remoteClusterCache{}
}

func (remoteClusterCache) AddCluster(remoteCluster remotecluster.RemoteCluster) {
	remoteClusterListCache.Store(remoteCluster.GetClusterID(), remoteCluster)
}

func (remoteClusterCache) DeleteCluster(clusterID string) {
	clusterID = strings.ToLower(clusterID)
	remoteClusterListCache.Delete(clusterID)
}

func (remoteClusterCache) GetCluster(clusterID string) (remotecluster.RemoteCluster, bool) {
	clusterID = strings.ToLower(clusterID)
	cluster, ok := remoteClusterListCache.Load(clusterID)
	if !ok {
		return nil, false
	}
	return cluster.(remotecluster.RemoteCluster), ok
}

func (remoteClusterCache) ListClusters() []remotecluster.RemoteCluster {
	list := []remotecluster.RemoteCluster{}
	remoteClusterListCache.Range(func(key, value interface{}) bool {
		list = append(list, value.(remotecluster.RemoteCluster))
		return true
	})
	return list
}

func (remoteClusterCache) UpdateCluster(clusterID string, remoteCluster remotecluster.RemoteCluster) (remotecluster.RemoteCluster, bool) {
	clusterID = strings.ToLower(clusterID)
	prevVal, ok := remoteClusterListCache.Swap(clusterID, remoteCluster)
	return prevVal.(remotecluster.RemoteCluster), ok
}

func (remoteClusterCache) Reset() {
	remoteClusterListCache = sync.Map{}
}
