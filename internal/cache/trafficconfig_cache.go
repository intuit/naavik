package cache

import (
	"strings"
	"sync"

	"github.com/intuit/naavik/internal/types/trafficconfig"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
)

var log = logger.NewLogger()

type TrafficConfigCacheInterface interface {
	BaseCache
	GetTrafficConfigEntry(identity string) *TrafficConfigEntry
	Get(identity string, env string) *admiralv1.TrafficConfig
	AddTrafficConfigToCache(trafficConfig *admiralv1.TrafficConfig)
	DeleteTrafficConfigFromCache(trafficConfig *admiralv1.TrafficConfig)
	GetTotalTrafficConfigs() int
}

var TrafficConfigCache = newTrafficConfigCache()

var trafficConfigCache = &trafficConfigCacheItem{
	cache: make(map[string]*TrafficConfigEntry),
}

type trafficConfigCacheReceiver struct{}

type trafficConfigCacheItem struct {
	cache map[string]*TrafficConfigEntry // map[Identity]*TrafficConfigEntry
	mutex sync.Mutex
}

type TrafficConfigEntry struct {
	sync.Mutex
	Identity               string                                       `json:"identity"`
	EnvTrafficConfig       map[string]*admiralv1.TrafficConfig          // map[env]*v1.TrafficConfig
	EnvServiceRoutesConfig map[string]*trafficconfig.ServiceRouteConfig // Map of env routes
}

func (tce *TrafficConfigEntry) Copy() *TrafficConfigEntry {
	tce.Lock()
	defer tce.Unlock()
	copyTce := &TrafficConfigEntry{
		Identity:               tce.Identity,
		EnvTrafficConfig:       make(map[string]*admiralv1.TrafficConfig, len(tce.EnvTrafficConfig)),
		EnvServiceRoutesConfig: make(map[string]*trafficconfig.ServiceRouteConfig, len(tce.EnvServiceRoutesConfig)),
	}
	for env, tc := range tce.EnvTrafficConfig {
		copyTce.EnvTrafficConfig[env] = tc.DeepCopy()
	}
	for env, src := range tce.EnvServiceRoutesConfig {
		copyTce.EnvServiceRoutesConfig[env] = src.Copy()
	}
	return copyTce
}

func newTrafficConfigCache() TrafficConfigCacheInterface {
	return &trafficConfigCacheReceiver{}
}

func (tcc *trafficConfigCacheReceiver) GetTrafficConfigEntry(identity string) *TrafficConfigEntry {
	identity = strings.ToLower(identity)
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	tce := trafficConfigCache.cache[identity]
	if tce == nil {
		return nil
	}
	// TODO: Evaluate using .Copy() to avoid lock contention
	return tce.Copy()
}

func (tcc *trafficConfigCacheReceiver) Get(identity string, env string) *admiralv1.TrafficConfig {
	identity = strings.ToLower(identity)
	env = strings.ToLower(env)
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	tce := trafficConfigCache.cache[identity]
	if tce != nil {
		tceenv := tce.EnvTrafficConfig[env]
		if tceenv == nil {
			return nil
		}
		return tceenv.DeepCopy()
	}
	return nil
}

func (tcc *trafficConfigCacheReceiver) AddTrafficConfigToCache(trafficConfig *admiralv1.TrafficConfig) {
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	tcutil := utils.TrafficConfigUtil(trafficConfig)
	identity := strings.ToLower(tcutil.GetIdentity())
	env := strings.ToLower(tcutil.GetEnv())
	if len(env) > 0 {
		tce := trafficConfigCache.cache[identity]
		if tce == nil {
			tce = &TrafficConfigEntry{
				Identity:               identity,
				EnvTrafficConfig:       make(map[string]*admiralv1.TrafficConfig),
				EnvServiceRoutesConfig: make(map[string]*trafficconfig.ServiceRouteConfig),
			}
		}
		tce.Lock()
		defer tce.Unlock()
		tce.EnvTrafficConfig[env] = trafficConfig
		tce.EnvServiceRoutesConfig[env] = getRoutesFromEdgeService(tcutil, trafficConfig)
		trafficConfigCache.cache[tce.Identity] = tce
	}
}

func (tcc *trafficConfigCacheReceiver) DeleteTrafficConfigFromCache(trafficConfig *admiralv1.TrafficConfig) {
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	tcutil := utils.TrafficConfigUtil(trafficConfig)
	identity := strings.ToLower(tcutil.GetIdentity())
	env := strings.ToLower(tcutil.GetEnv())
	if len(env) > 0 {
		tce := trafficConfigCache.cache[identity]
		if tce != nil {
			tce.Lock()
			defer tce.Unlock()
			if tce.EnvTrafficConfig[env] != nil && trafficConfig.Name == tce.EnvTrafficConfig[env].Name {
				log.Str(logger.OperationKey, logger.DeleteValue).
					Str(logger.ResourceIdentifierKey, trafficConfig.Name).
					Str(logger.NamespaceKey, trafficConfig.Namespace).
					Trace("ignoring trafficconfig and deleting from cache")
				delete(tce.EnvTrafficConfig, env)
				if tce.EnvServiceRoutesConfig[env] != nil {
					delete(tce.EnvServiceRoutesConfig, env)
				}
			}
		}
	}
}

func (tcc *trafficConfigCacheReceiver) GetTotalTrafficConfigs() int {
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	counter := 0
	for _, tce := range trafficConfigCache.cache {
		counter += len(tce.EnvTrafficConfig)
	}
	return counter
}

func getRoutesFromEdgeService(tcutil utils.TrafficConfigInterface, tc *admiralv1.TrafficConfig) *trafficconfig.ServiceRouteConfig {
	src := &trafficconfig.ServiceRouteConfig{WorkloadEnvRevision: map[string]string{}}
	routes := []*admiralv1.Route{}
	if tc.Spec.EdgeService != nil {
		for _, route := range tc.Spec.EdgeService.Routes {
			if len(route.WorkloadEnvSelectors) > 0 {
				routeItem := &admiralv1.Route{}
				routeItem.Name = route.Name
				routeItem.Inbound = route.Inbound
				routeItem.Outbound = route.Outbound
				routeItem.WorkloadEnvSelectors = route.WorkloadEnvSelectors
				routes = append(routes, routeItem)
			}
		}
		src.Routes = routes
		src.ServiceAssetAlias = tcutil.GetIdentity()
		for _, env := range tc.Spec.WorkloadEnv {
			src.WorkloadEnvRevision[strings.ToLower(env)] = tcutil.GetRevision()
		}
	}
	return src
}

func (trafficConfigCacheReceiver) Reset() {
	defer trafficConfigCache.mutex.Unlock()
	trafficConfigCache.mutex.Lock()
	trafficConfigCache.cache = make(map[string]*TrafficConfigEntry)
}
