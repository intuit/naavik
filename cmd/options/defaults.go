package options

import (
	"time"

	"github.com/intuit/naavik/internal/types"
)

const (
	DefaultSecretNamespace            = "admiral"
	DefaultDependencyNamespace        = "admiral"
	DefaultClusterRegistriesNamespace = "admiral"
	DefaultSyncNamespace              = "admiral-sync"
	DefaultWorkloadIdentity           = "alpha.istio.io/identity"
	DefaultWorkloadEnvKey             = "admiral.io/env"
	DefaultResourceIgnoreLabel        = "admiral.io/ignore"
	DefaultSecretSyncLabel            = "admiral.io/sync"
	DefaultMeshInjectionKey           = "sidecar.istio.io/inject"
	DefaultHostnameSuffix             = "mesh"
	DefaultConfigPath                 = "/etc/admiral/config.yaml"
	DefaultLogLevel                   = "info"
	DefaultLogColor                   = false
	DefaultProfilerEndpoint           = "localhost:4040"
	DefaultArgoRolloutsEnabled        = true
	DefaultEnableProfiling            = false
	DefaultConfigResolver             = types.ConfigResolverSecret
	DefaultStateChecker               = types.StateCheckerNone
	DefaultTrafficConfigNamespace     = "admiral"
	DefaultTrafficConfigIdentityKey   = "asset"
	DefaultRefreshInterval            = time.Minute
	DefaultAsyncExecutorMaxGoRoutines = 20000
	DefaultWorkerConcurrency          = 1
)

var (
	DefaultTrafficConfigClustersScope    = []string{".*"}
	DefaultIgnoreAssetAliases            = []string{}
	DefaultEnvoyFilterVersions           = []string{"1.21"}
	DefaultDeprecatedEnvoyFilterVersions = []string{"1.13"}
	DefaultDisabledFeatures              = []string{""}

	AvailableFeatures = []types.FeatureName{types.FeatureThrottleFilter, types.FeatureVirtualService}
)
