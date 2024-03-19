package cmd

import (
	"flag"
	"fmt"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/bootstrap"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "naavik",
	Short: "Naavik is a control plane to manage Intuit features of Istio Service Mesh",
	Long:  `Naavik is a control plane to manage Intuit features of Istio Service Mesh based on unified configuration model.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.NewContextWithLogger()
		bootstrap.InitNaavik(ctx)
	},
}

// Execute executes the root command.
func Execute() error {
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	return rootCmd.Execute()
}

func init() {
	// TODO: Add support for config file
	// Basic options
	rootCmd.PersistentFlags().StringVar(&options.Params.LogLevel, "log_level", options.DefaultLogLevel,
		fmt.Sprintf("Set log verbosity, defaults to 'Info'. Must be between %q and %q", "trace", "info"))
	rootCmd.PersistentFlags().BoolVar(&options.Params.LogColor, "log_color", options.DefaultLogColor,
		fmt.Sprintf("Enable color for logs. Default is %t", options.DefaultLogColor))
	rootCmd.PersistentFlags().StringVar(&options.Params.ConfigResolver, "config_resolver", options.DefaultConfigResolver,
		fmt.Sprintf("Set the config resolver to run naavik with, defaults to %q", options.DefaultConfigResolver))
	rootCmd.PersistentFlags().StringVar(&options.Params.StateChecker, "state_checker", options.DefaultStateChecker,
		fmt.Sprintf("Set the state checker to run naavik with, defaults to %q", options.DefaultStateChecker))
	rootCmd.PersistentFlags().StringVar(&options.Params.ConfigPath, "config_path", options.DefaultConfigPath,
		fmt.Sprintf("Path where the configuration file resides. Defaults to %q", options.DefaultConfigPath))
	rootCmd.PersistentFlags().BoolVar(&options.Params.EnableProfiling, "enable_profiling", options.DefaultEnableProfiling,
		fmt.Sprintf("Enable go profiling for cpu, memory, goroutines, etc. Defaults to %t", options.DefaultEnableProfiling))
	rootCmd.PersistentFlags().StringVar(&options.Params.ProfilerEndpoint, "profiler_endpoint", options.DefaultProfilerEndpoint,
		fmt.Sprintf("Set the continuous profiler endpoint. Defaults to %q", options.DefaultProfilerEndpoint))

	// Controller options
	rootCmd.PersistentFlags().BoolVar(&options.Params.ArgoRolloutsEnabled, "argo_rollouts", options.DefaultArgoRolloutsEnabled,
		fmt.Sprintf("Use argo rollout configurations. Defaults to %t", options.DefaultArgoRolloutsEnabled))
	rootCmd.PersistentFlags().IntVar(&options.Params.WorkerConcurrency, "worker_concurrency", options.DefaultWorkerConcurrency,
		fmt.Sprintf("Number of workers to process events from informers (This is per controller config). Defaults to %d", options.DefaultWorkerConcurrency))
	rootCmd.PersistentFlags().StringVar(&options.Params.KubeConfigPath, "kube_config", "",
		"Use a Kubernetes configuration file instead of in-cluster configuration. Defaults to empty string, which means in-cluster configuration")
	rootCmd.PersistentFlags().StringVar(&options.Params.SyncNamespace, "sync_namespace", options.DefaultSyncNamespace,
		fmt.Sprintf("Namespace to monitor for custom resources. Defaults to %q", options.DefaultSyncNamespace))
	rootCmd.PersistentFlags().StringVar(&options.Params.ClusterRegistriesNamespace, "secret_namespace", options.DefaultSecretNamespace,
		fmt.Sprintf("Namespace to monitor for secrets that contains remote cluster data. Defaults to %q", options.DefaultSecretNamespace))
	rootCmd.PersistentFlags().StringVar(&options.Params.DependenciesNamespace, "dependency_namespace", options.DefaultDependencyNamespace,
		fmt.Sprintf("Namespace to monitor for service dependency data. Defaults to %q", options.DefaultDependencyNamespace))
	rootCmd.PersistentFlags().DurationVar(&options.Params.CacheRefreshInterval, "sync_period", options.DefaultRefreshInterval,
		fmt.Sprintf("Interval for syncing Kubernetes resources. Defaults to %d", options.DefaultRefreshInterval))
	rootCmd.PersistentFlags().IntVar(&options.Params.AsyncExecutorMaxGoRoutines, "async_executor_max_goroutines", options.DefaultAsyncExecutorMaxGoRoutines,
		fmt.Sprintf("Maximum number of go routines to be used by async executor. Defaults to %d", options.DefaultAsyncExecutorMaxGoRoutines))

	// Workload options
	rootCmd.PersistentFlags().StringVar(&options.Params.WorkloadIdentityKey, "workload_identity_key", options.DefaultWorkloadIdentity,
		fmt.Sprintf("The workload identity  key, on deployment/rollout which holds identity value used to generate cname. Default label key will be %q"+
			"If present, that will be used. If not, it will try an annotation (for use cases where an identity is longer than 63 chars)", options.DefaultWorkloadIdentity),
	)
	rootCmd.PersistentFlags().StringVar(&options.Params.EnvKey, "env_key", options.DefaultWorkloadEnvKey,
		fmt.Sprintf("The annotation or label, on a pod spec in a deployment/rollout, which will be used to group deployments across regions/clusters under a single environment. Defaults to %q"+
			"The order would be to use annotation specified as `env_key`, followed by label specified as `env_key` and then fallback to the label `env`", options.DefaultWorkloadEnvKey))
	rootCmd.PersistentFlags().StringVar(&options.Params.ResourceIgnoreLabel, "resource_ignore_label", options.DefaultResourceIgnoreLabel,
		fmt.Sprintf("The label on the resource, which will be used to ignore the resource from getting processed. Defaults to %q", options.DefaultResourceIgnoreLabel))
	rootCmd.PersistentFlags().StringVar(&options.Params.SecretSyncLabel, "secret_sync_label", options.DefaultSecretSyncLabel,
		fmt.Sprintf("The label on the secret, which will be used to sync the secret of remote clusters. Defaults to %q", options.DefaultSecretSyncLabel))
	rootCmd.PersistentFlags().StringVar(&options.Params.HostnameSuffix, "hostname_suffix", options.DefaultHostnameSuffix,
		fmt.Sprintf("The hostname suffix to customize the cname generated by admiral. Default suffix value will be %q", options.DefaultHostnameSuffix))
	rootCmd.PersistentFlags().StringVar(&options.Params.MeshInjectionEnabledKey, "injection_enabled_label_key", options.DefaultMeshInjectionKey,
		fmt.Sprintf("The hostname suffix to customize the cname generated by admiral. Default suffix value will be %q", options.DefaultMeshInjectionKey))

	// Traffic options
	rootCmd.PersistentFlags().
		StringVar(&options.Params.TrafficConfigNamespace, "traffic_config_namespace", options.DefaultTrafficConfigNamespace,
			fmt.Sprintf("Namespace to monitor for service traffic config data. Defaults to %q", options.DefaultTrafficConfigNamespace))
	rootCmd.PersistentFlags().StringVar(&options.Params.TrafficConfigIdentityKey, "traffic_config_identity_key", options.DefaultTrafficConfigIdentityKey,
		fmt.Sprintf("The traffic config identity key holds identity value of a service. Default label key will be %q.", options.DefaultTrafficConfigIdentityKey))
	rootCmd.PersistentFlags().StringArrayVar(&options.Params.AllowedClusterScope, "traffic_config_clusters_scope", options.DefaultTrafficConfigClustersScope,
		fmt.Sprintf("List of clusters that should be processed for traffic config. Defaults to %q", options.DefaultTrafficConfigClustersScope))
	rootCmd.PersistentFlags().StringArrayVar(&options.Params.IgnoreAssetAliases, "ignore_asset_aliases", options.DefaultIgnoreAssetAliases,
		fmt.Sprintf("List of asset aliases that should be ignored for traffic config processing. Defaults to %q", options.DefaultIgnoreAssetAliases))
	rootCmd.PersistentFlags().StringArrayVar(&options.Params.EnvoyFilterVersions, "envoy_filter_versions", options.DefaultEnvoyFilterVersions,
		fmt.Sprintf("List of envoy filter versions that should be processed for traffic config. Defaults to %q", options.DefaultEnvoyFilterVersions))
	rootCmd.PersistentFlags().StringArrayVar(&options.Params.DeprecatedEnvoyFilterVersions, "deprecated_envoy_filter_versions", options.DefaultDeprecatedEnvoyFilterVersions,
		fmt.Sprintf("List of envoy filter versions that are deprecated and should be removed while traffic config processing. Defaults to %q", options.DefaultDeprecatedEnvoyFilterVersions))
	rootCmd.PersistentFlags().StringArrayVar(&options.Params.DisabledFeatures, "disabled_features", options.DefaultDisabledFeatures,
		fmt.Sprintf("Comma separated list of features to be disabled. Available features %v", options.AvailableFeatures))
}
