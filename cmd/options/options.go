package options

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/intuit/naavik/internal/types"
)

var (
	Params      = NaavikArgs{}
	StartUpTime = time.Now()
)

// NaavikArgs contains the arguments for the naavik.
type NaavikArgs struct {
	LogLevel                string
	LogColor                bool
	ArgoRolloutsEnabled     bool
	ConfigPath              string
	WorkloadIdentityKey     string
	MeshInjectionEnabledKey string
	EnvKey                  string
	ResourceIgnoreLabel     string
	SecretSyncLabel         string
	HostnameSuffix          string
	EnableProfiling         bool
	ProfilerEndpoint        string

	ConfigResolver string
	StateChecker   string

	TrafficConfigNamespace        string
	TrafficConfigIdentityKey      string
	AllowedClusterScope           []string
	IgnoreAssetAliases            []string
	EnvoyFilterVersions           []string
	DeprecatedEnvoyFilterVersions []string
	DisabledFeatures              []string
	AsyncExecutorMaxGoRoutines    int
	WorkerConcurrency             int

	CacheRefreshInterval time.Duration

	KubeConfigPath             string
	ClusterRegistriesNamespace string
	DependenciesNamespace      string
	SyncNamespace              string
}

func GetNaavikArgs() NaavikArgs {
	return Params
}

func GetLogLevel() string {
	return Params.LogLevel
}

func GetLogColor() bool {
	return Params.LogColor
}

func IsArgoRolloutsEnabled() bool {
	return Params.ArgoRolloutsEnabled
}

func IsProfilingEnabled() bool {
	return Params.EnableProfiling
}

func GetProfilerEndpoint() string {
	return Params.ProfilerEndpoint
}

func GetConfigPath() string {
	return Params.ConfigPath
}

func GetWorkloadIdentityKey() string {
	return Params.WorkloadIdentityKey
}

func GetMeshInjectionEnabledKey() string {
	return Params.MeshInjectionEnabledKey
}

func GetEnvKey() string {
	return Params.EnvKey
}

func GetResourceIgnoreLabel() string {
	return Params.ResourceIgnoreLabel
}

func GetSecretSyncLabel() string {
	return Params.SecretSyncLabel
}

func GetHostnameSuffix() string {
	return Params.HostnameSuffix
}

func GetTrafficConfigNamespace() string {
	return Params.TrafficConfigNamespace
}

func GetTrafficConfigIdentityKey() string {
	return Params.TrafficConfigIdentityKey
}

func GetTrafficConfigScope() []string {
	return Params.AllowedClusterScope
}

func GetStateChecker() string {
	return Params.StateChecker
}

func GetConfigResolver() string {
	return Params.ConfigResolver
}

func IsClusterInAllowedScope(cluster string) bool {
	for _, scope := range Params.AllowedClusterScope {
		r, e := regexp.Compile("(?i)" + strings.TrimSpace(scope))
		if e == nil {
			if r.MatchString(cluster) {
				return true
			}
		} else {
			panic(fmt.Errorf("unable to compile regex for scope=%s", scope))
		}
	}
	return false
}

func GetTrafficConfigIgnoreAssets() []string {
	assets := []string{}
	for _, asset := range Params.IgnoreAssetAliases {
		assets = append(assets, strings.ToLower(asset))
	}
	return assets
}

func IsAssetIgnored(asset string) bool {
	return slices.Contains(GetTrafficConfigIgnoreAssets(), strings.ToLower(asset))
}

func GetEnvoyFilterVersions() []string {
	return Params.EnvoyFilterVersions
}

func GetDeprecatedEnvoyFilterVersions() []string {
	return Params.DeprecatedEnvoyFilterVersions
}

func GetAsyncExecutorMaxGoRoutines() int32 {
	return int32(Params.AsyncExecutorMaxGoRoutines)
}

func GetWorkerConcurrency() int {
	return Params.WorkerConcurrency
}

func GetCacheRefreshInterval() time.Duration {
	return Params.CacheRefreshInterval
}

func GetKubeConfigPath() string {
	return Params.KubeConfigPath
}

func GetClusterRegistriesNamespace() string {
	return Params.ClusterRegistriesNamespace
}

func GetDependenciesNamespace() string {
	return Params.DependenciesNamespace
}

func GetSyncNamespace() string {
	return Params.SyncNamespace
}

func GetStartUpTime() time.Time {
	return StartUpTime
}

func IsCacheWarmedUp() bool {
	return time.Since(StartUpTime) > Params.CacheRefreshInterval
}

func GetEnvironment() string {
	env := os.Getenv(types.AppEnvKey)
	if len(env) == 0 {
		return types.EnvDev
	}
	return env
}

func IsFeatureEnabled(feature types.FeatureName) bool {
	return !slices.Contains(Params.DisabledFeatures, feature.String())
}

func InitializeNaavikArgs(args *NaavikArgs) {
	if args == nil {
		args = &NaavikArgs{}
	}

	Params = NaavikArgs{
		LogLevel:                      getValueOrDefault[string](args.LogLevel, DefaultLogLevel),
		ArgoRolloutsEnabled:           getValueOrDefault[bool](args.ArgoRolloutsEnabled, DefaultArgoRolloutsEnabled),
		ConfigResolver:                getValueOrDefault[string](args.ConfigResolver, DefaultConfigResolver),
		StateChecker:                  getValueOrDefault[string](args.StateChecker, DefaultStateChecker),
		ConfigPath:                    getValueOrDefault[string](args.ConfigPath, DefaultConfigPath),
		WorkloadIdentityKey:           getValueOrDefault[string](args.WorkloadIdentityKey, DefaultWorkloadIdentity),
		EnvKey:                        getValueOrDefault[string](args.EnvKey, DefaultWorkloadEnvKey),
		ResourceIgnoreLabel:           getValueOrDefault[string](args.ResourceIgnoreLabel, DefaultResourceIgnoreLabel),
		SecretSyncLabel:               getValueOrDefault[string](args.SecretSyncLabel, DefaultSecretSyncLabel),
		HostnameSuffix:                getValueOrDefault[string](args.HostnameSuffix, DefaultHostnameSuffix),
		EnableProfiling:               getValueOrDefault[bool](args.EnableProfiling, DefaultEnableProfiling),
		TrafficConfigNamespace:        getValueOrDefault[string](args.TrafficConfigNamespace, DefaultTrafficConfigNamespace),
		TrafficConfigIdentityKey:      getValueOrDefault[string](args.TrafficConfigIdentityKey, DefaultTrafficConfigIdentityKey),
		AllowedClusterScope:           getValueOrDefaultSlice(args.AllowedClusterScope, DefaultTrafficConfigClustersScope),
		IgnoreAssetAliases:            getValueOrDefaultSlice(args.IgnoreAssetAliases, DefaultIgnoreAssetAliases),
		EnvoyFilterVersions:           getValueOrDefaultSlice(args.EnvoyFilterVersions, DefaultEnvoyFilterVersions),
		DeprecatedEnvoyFilterVersions: getValueOrDefaultSlice(args.DeprecatedEnvoyFilterVersions, DefaultDeprecatedEnvoyFilterVersions),
		DisabledFeatures:              getValueOrDefaultSlice(args.DisabledFeatures, DefaultDisabledFeatures),
		AsyncExecutorMaxGoRoutines:    getValueOrDefault[int](args.AsyncExecutorMaxGoRoutines, DefaultAsyncExecutorMaxGoRoutines),
		MeshInjectionEnabledKey:       getValueOrDefault[string](args.MeshInjectionEnabledKey, DefaultMeshInjectionKey),
		WorkerConcurrency:             getValueOrDefault[int](args.WorkerConcurrency, DefaultWorkerConcurrency),
		KubeConfigPath:                getValueOrDefault[string](args.ClusterRegistriesNamespace, ""),
		ClusterRegistriesNamespace:    getValueOrDefault[string](args.ClusterRegistriesNamespace, DefaultClusterRegistriesNamespace),
		DependenciesNamespace:         getValueOrDefault[string](args.DependenciesNamespace, DefaultDependencyNamespace),
		SyncNamespace:                 getValueOrDefault[string](args.SyncNamespace, DefaultSyncNamespace),
		CacheRefreshInterval:          getValueOrDefault[time.Duration](args.CacheRefreshInterval, DefaultRefreshInterval),
	}
}

func getValueOrDefaultSlice(val, def []string) []string {
	if val == nil {
		return def
	}
	return val
}

func getValueOrDefault[T comparable](val, def T) T {
	var zeroVal T
	if zeroVal == val {
		return def
	}
	return val
}
