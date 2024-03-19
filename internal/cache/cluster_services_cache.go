package cache

import (
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
)

type ServicesCache interface {
	BaseCache
	GetByClusterNamespace(clusterID, identity string) *K8sServiceEntry
	GetByNamespace(identity string) []*K8sServiceItem
	Add(clusterID string, service *corev1.Service)
	Delete(clusterID string, service *corev1.Service)
}

var (
	servicecache     = sync.Map{} // map[clusterID]serviceIdentityMap
	servicecacheLock = sync.RWMutex{}
)

type (
	serviceIdentityMapEntry = map[string]*K8sServiceEntry // map[namespace]*K8sServiceEntry
)

type K8sServiceEntry struct {
	Namespace string                     `json:"namespace"`
	Services  map[string]*K8sServiceItem // map[env]*K8sServiceItem
}

func (kse *K8sServiceEntry) Copy() *K8sServiceEntry {
	kseCopy := K8sServiceEntry{
		Namespace: kse.Namespace,
		Services:  map[string]*K8sServiceItem{},
	}
	for k, v := range kse.Services {
		kseCopy.Services[k] = &K8sServiceItem{
			Service:   v.Service.DeepCopy(),
			ClusterID: v.ClusterID,
		}
	}
	return &kseCopy
}

type K8sServiceItem struct {
	Service   *corev1.Service `json:"service"`
	ClusterID string          `json:"clusterId"`
}

var Services = newServicesCache()

type serviceClusterCache struct{}

func newServicesCache() ServicesCache {
	return &serviceClusterCache{}
}

func (serviceClusterCache) GetByClusterNamespace(clusterID, namespace string) *K8sServiceEntry {
	clusterID = strings.ToLower(clusterID)
	namespace = strings.ToLower(namespace)
	servicecacheLock.RLock()
	defer servicecacheLock.RUnlock()
	serviceIdentityMapEntryVal, ok := servicecache.Load(clusterID)
	if !ok {
		return nil
	}
	serviceIdentityMap := serviceIdentityMapEntryVal.(serviceIdentityMapEntry)
	serviceEntry, ok := serviceIdentityMap[namespace]
	if !ok {
		return nil
	}
	return serviceEntry.Copy()
}

func (serviceClusterCache) GetByNamespace(namespace string) []*K8sServiceItem {
	namespace = strings.ToLower(namespace)
	list := []*K8sServiceItem{}
	servicecacheLock.RLock()
	defer servicecacheLock.RUnlock()
	servicecache.Range(func(key, value interface{}) bool {
		serviceIdentityMap := value.(serviceIdentityMapEntry)
		serviceEntry, ok := serviceIdentityMap[namespace]
		if !ok {
			return true
		}
		for _, serviceItem := range serviceEntry.Services {
			copyServiceItem := &K8sServiceItem{
				Service:   serviceItem.Service.DeepCopy(),
				ClusterID: serviceItem.ClusterID,
			}
			list = append(list, copyServiceItem)
		}
		return true
	})
	return list
}

func (serviceClusterCache) Add(clusterID string, service *corev1.Service) {
	clusterID = strings.ToLower(clusterID)
	namespace := service.Namespace
	svcName := service.Name
	servicecacheLock.Lock()
	defer servicecacheLock.Unlock()
	serviceIdentityMapEntryVal, ok := servicecache.Load(clusterID)
	if !ok {
		serviceIdentityMap := serviceIdentityMapEntry{}
		serviceIdentityMap[namespace] = &K8sServiceEntry{
			Namespace: namespace,
			Services: map[string]*K8sServiceItem{
				svcName: {
					Service:   service,
					ClusterID: clusterID,
				},
			},
		}
		servicecache.Store(clusterID, serviceIdentityMap)
		return
	}
	serviceIdentityMap := serviceIdentityMapEntryVal.(serviceIdentityMapEntry)
	serviceEntry, ok := serviceIdentityMap[namespace]
	if !ok {
		serviceIdentityMap[namespace] = &K8sServiceEntry{
			Namespace: namespace,
			Services: map[string]*K8sServiceItem{
				svcName: {
					Service:   service,
					ClusterID: clusterID,
				},
			},
		}
		return
	}
	serviceEntry.Services[svcName] = &K8sServiceItem{
		Service:   service,
		ClusterID: clusterID,
	}
}

func (serviceClusterCache) Delete(clusterID string, service *corev1.Service) {
	clusterID = strings.ToLower(clusterID)
	namespace := service.Namespace
	svcName := service.Name
	servicecacheLock.Lock()
	defer servicecacheLock.Unlock()
	serviceIdentityMapEntryVal, ok := servicecache.Load(clusterID)
	if !ok {
		return
	}
	serviceIdentityMap := serviceIdentityMapEntryVal.(serviceIdentityMapEntry)
	serviceEntry, ok := serviceIdentityMap[namespace]
	if !ok {
		return
	}
	delete(serviceEntry.Services, svcName)
	if len(serviceEntry.Services) == 0 {
		delete(serviceIdentityMap, namespace)
	}
	if len(serviceIdentityMap) == 0 {
		servicecache.Delete(clusterID)
	}
}

func (serviceClusterCache) Reset() {
	servicecacheLock.Lock()
	defer servicecacheLock.Unlock()
	servicecache = sync.Map{}
}
