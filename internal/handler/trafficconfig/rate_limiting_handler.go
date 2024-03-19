package trafficconfig

import (
	"fmt"
	"strings"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	localratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/internal/types/remotecluster"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"istio.io/api/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var globalBucket = getTokenBucket(1000000, 1000000, time.Second)

func HandleRateLimiter(ctx context.Context, trafficConfig *admiralv1.TrafficConfig, eventType types.EventType) {
	tcUtil := utils.TrafficConfigUtil(trafficConfig)

	clusters := cache.IdentityCluster.GetClustersForIdentity(tcUtil.GetIdentity())

	if len(clusters) == 0 {
		ctx.Log.Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Warn("no clusters found for identity.")
		return
	}
	ctx.Log.Any(logger.ClusterKey, clusters).Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Info("clusters for the identity")

	for _, clusterID := range clusters {
		if !options.IsClusterInAllowedScope(clusterID) {
			ctx.Log.Str(logger.ClusterKey, clusterID).Warn("cluster is not in allowed scope. skipping thorttle filter processing.")
			continue
		}
		rc, found := cache.RemoteCluster.GetCluster(clusterID)
		if !found {
			ctx.Log.Str(logger.ClusterKey, clusterID).Error("cluster not found in cache. skipping thorttle filter processing.")
			continue
		}

		if tcUtil.IsDisabled() || eventType == types.Delete {
			filterList, err := listRateLimitingFilters(ctx, rc, tcUtil)
			if err != nil {
				ctx.Log.Str(logger.ClusterKey, rc.GetClusterID()).Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Str(logger.EnvKey, tcUtil.GetEnv()).Str(logger.ErrorKey, err.Error()).Warn("failed to list envoy filters for identity with Latest LabelSet")
			}
			if err := rc.IstioClient().DeleteEnvoyFilters(ctx, filterList.Items); err != nil {
				ctx.Log.Str(logger.ClusterKey, rc.GetClusterID()).Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Str(logger.EnvKey, tcUtil.GetEnv()).Str(logger.ErrorKey, err.Error()).Warn("failed to delete envoy filters for identity with Latest LabelSet")
			}
			continue
		}

		createRateLimitingFilters(ctx, rc, tcUtil)
	}
}

func createRateLimitingFilters(ctx context.Context, rc remotecluster.RemoteCluster, tcUtil utils.TrafficConfigInterface) {
	newList := make([]*networkingv1alpha3.EnvoyFilter, 0)

	for _, env := range tcUtil.GetWorkloadEnvs() {
		workloadLabels, err := getWorkLoadLabels(ctx, rc.GetClusterID(), tcUtil.GetIdentity(), env)
		if err != nil {
			ctx.Log.Str(logger.ClusterKey, rc.GetClusterID()).Warnf("skipping, %s", err.Error())
			continue
		}

		for _, version := range options.GetEnvoyFilterVersions() {
			envoyFilterName := utils.EnvoyFilterUtil().GetName(tcUtil.GetIdentity(), "throttle", env+"-"+version)

			envoyFilter := &networkingv1alpha3.EnvoyFilter{
				TypeMeta: metav1.TypeMeta{
					Kind:       "EnvoyFilter",
					APIVersion: "networking.istio.io/v1alpha3",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      envoyFilterName,
					Namespace: types.NamespaceIstioSystem,
					Annotations: map[string]string{
						types.RevisionNumberKey: tcUtil.GetRevision(),
						types.TransactionIDKey:  tcUtil.GetTransactionID(),
					},
					Labels: map[string]string{
						types.CreatedForKey:           strings.ToLower(tcUtil.GetIdentity()),
						types.CreatedByKey:            types.NaavikName,
						types.CreatedForEnvKey:        env,
						types.CreatedTypeKey:          "throttle_filter",
						types.CreatedForTrafficEnvKey: tcUtil.GetEnv(),
					},
				},
				Spec: v1alpha3.EnvoyFilter{
					Priority:         0,
					WorkloadSelector: &v1alpha3.WorkloadSelector{Labels: workloadLabels},
					ConfigPatches:    createConfigPatches(env, version, tcUtil, rc),
				},
			}

			f, _ := rc.IstioClient().GetEnvoyFilter(ctx, envoyFilterName, types.NamespaceIstioSystem, metav1.GetOptions{})
			if f != nil {
				envoyFilter.ResourceVersion = f.ResourceVersion
			}

			newList = append(newList, envoyFilter)
		}
	}

	oldList, err := listRateLimitingFilters(ctx, rc, tcUtil)
	if err != nil {
		ctx.Log.Str(logger.ClusterKey, rc.GetClusterID()).Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Str(logger.EnvKey, tcUtil.GetEnv()).
			Str(logger.ErrorKey, err.Error()).Warn("failed to list envoy filters for identity with Latest LabelSet")
	}

	rc.IstioClient().ApplyEnvoyFilters(ctx, newList, oldList)
}

func createConfigPatches(env, proxyVersion string, tcUtil utils.TrafficConfigInterface, rc remotecluster.RemoteCluster) []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch {
	patches := []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{createFilterPatch(proxyVersion)}
	patches = append(patches, createRoutePatches(env, tcUtil, rc)...)
	return patches
}

func createFilterPatch(proxyVersion string) *v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch {
	return &v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
		ApplyTo: v1alpha3.EnvoyFilter_HTTP_FILTER,
		Match: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
			Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
			Proxy:   createProxyMatch(proxyVersion),
			ObjectTypes: &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_Listener{
				Listener: &v1alpha3.EnvoyFilter_ListenerMatch{
					FilterChain: &v1alpha3.EnvoyFilter_ListenerMatch_FilterChainMatch{
						Filter: &v1alpha3.EnvoyFilter_ListenerMatch_FilterMatch{
							Name: "envoy.filters.network.http_connection_manager",
							SubFilter: &v1alpha3.EnvoyFilter_ListenerMatch_SubFilterMatch{
								Name: "envoy.filters.http.router",
							},
						},
					},
				},
			},
		},
		Patch: &v1alpha3.EnvoyFilter_Patch{
			Operation: v1alpha3.EnvoyFilter_Patch_INSERT_BEFORE,
			Value: &structpb.Struct{Fields: map[string]*structpb.Value{
				"name": {Kind: &structpb.Value_StringValue{StringValue: "envoy.filters.http.local_ratelimit"}},
				"typed_config": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"@type":    {Kind: &structpb.Value_StringValue{StringValue: "type.googleapis.com/udpa.type.v1.TypedStruct"}},
						"type_url": {Kind: &structpb.Value_StringValue{StringValue: "type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit"}},
						"value": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{Fields: map[string]*structpb.Value{
							"stat_prefix": {Kind: &structpb.Value_StringValue{StringValue: "http_local_rate_limiter"}},
						}}}},
					},
				}}},
			}},
		},
	}
}

func createRoutePatches(env string, tcUtil utils.TrafficConfigInterface, rc remotecluster.RemoteCluster) []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch {
	rateLimits := &structpb.ListValue{}
	descriptors := &structpb.ListValue{}

	for _, tcg := range tcUtil.GetQuotaGroup().TotalQuotaGroup {
		if !contains(tcg.WorkloadEnvSelectors, env) {
			continue
		}

		for _, quota := range tcg.Quotas {
			timePeriod, err := time.ParseDuration(quota.TimePeriod)
			if err != nil {
				fmt.Printf("error parsing time period %q for total quota %q : %+v", quota.TimePeriod, quota.Name, err)
				continue
			}

			descriptor := getDescriptorConfig(tcg.Name, "", quota, timePeriod)
			descriptorValue := structpb.NewStructValue(getProtoStructFromProtoMessage(descriptor))
			descriptors.Values = append(descriptors.Values, descriptorValue)

			rateLimit := getRateLimitConfig(tcg.Name, "", quota)
			rateLimitValue := structpb.NewStructValue(getProtoStructFromProtoMessage(rateLimit))
			rateLimits.Values = append(rateLimits.Values, rateLimitValue)
		}
	}

	for _, aqg := range tcUtil.GetQuotaGroup().AppQuotaGroups {
		if !contains(aqg.WorkloadEnvSelectors, env) {
			continue
		}

		for _, associatedApp := range aqg.AssociatedApps {
			for _, quota := range aqg.Quotas {
				timePeriod, err := time.ParseDuration(quota.TimePeriod)
				if err != nil {
					fmt.Printf("error parsing time period %q for app quota %q : %+v", quota.TimePeriod, quota.Name, err)
					continue
				}

				descriptor := getDescriptorConfig(aqg.Name, associatedApp, quota, timePeriod)
				descriptorValue := structpb.NewStructValue(getProtoStructFromProtoMessage(descriptor))
				descriptors.Values = append(descriptors.Values, descriptorValue)

				rateLimit := getRateLimitConfig(aqg.Name, associatedApp, quota)
				rateLimitValue := structpb.NewStructValue(getProtoStructFromProtoMessage(rateLimit))
				rateLimits.Values = append(rateLimits.Values, rateLimitValue)
			}
		}
	}

	routePatch := createRoutePatch(rateLimits, descriptors)

	inboundPorts := getInboundPorts(rc.GetClusterID(), tcUtil.GetIdentity(), env)
	routePatches := []*v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{}
	for _, inboundPort := range inboundPorts {
		routePatches = append(routePatches, &v1alpha3.EnvoyFilter_EnvoyConfigObjectPatch{
			ApplyTo: v1alpha3.EnvoyFilter_HTTP_ROUTE,
			Match:   createRouteMatch(inboundPort),
			Patch:   routePatch,
		})
	}
	return routePatches
}

func getInboundPorts(clusterID, identity, env string) []string {
	deploy := cache.Deployments.GetByClusterIdentityEnv(clusterID, identity, env)
	if deploy != nil {
		inboundPorts := deploy.Spec.Template.Annotations[types.IncludeInboundPortsAnnotation]
		return strings.Split(inboundPorts, ",")
	}

	rollout := cache.Rollouts.GetByClusterIdentityEnv(clusterID, identity, env)
	if rollout != nil {
		inboundPorts := rollout.Spec.Template.Annotations[types.IncludeInboundPortsAnnotation]
		return strings.Split(inboundPorts, ",")
	}

	return nil
}

func createProxyMatch(proxyVersion string) *v1alpha3.EnvoyFilter_ProxyMatch {
	return &v1alpha3.EnvoyFilter_ProxyMatch{
		ProxyVersion: "^" + strings.ReplaceAll(proxyVersion, ".", "\\.") + ".*",
	}
}

func createRouteMatch(inboundPort string) *v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch {
	if inboundPort == "" {
		return &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
			Context: v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
		}
	}

	return &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch{
		Context:     v1alpha3.EnvoyFilter_SIDECAR_INBOUND,
		ObjectTypes: getRouteConfig(inboundPort),
	}
}

func getRouteConfig(inboundPort string) *v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_RouteConfiguration {
	return &v1alpha3.EnvoyFilter_EnvoyConfigObjectMatch_RouteConfiguration{
		RouteConfiguration: &v1alpha3.EnvoyFilter_RouteConfigurationMatch{
			Vhost: &v1alpha3.EnvoyFilter_RouteConfigurationMatch_VirtualHostMatch{
				Name: fmt.Sprintf("inbound|http|%s", inboundPort),
				Route: &v1alpha3.EnvoyFilter_RouteConfigurationMatch_RouteMatch{
					Action: v1alpha3.EnvoyFilter_RouteConfigurationMatch_RouteMatch_ANY,
				},
			},
		},
	}
}

func createRoutePatch(rateLimits, descriptors *structpb.ListValue) *v1alpha3.EnvoyFilter_Patch {
	return &v1alpha3.EnvoyFilter_Patch{
		Operation: v1alpha3.EnvoyFilter_Patch_MERGE,
		Value: &structpb.Struct{Fields: map[string]*structpb.Value{
			"route": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"rate_limits": structpb.NewListValue(rateLimits),
				},
			}),
			"typed_per_filter_config": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"envoy.filters.http.local_ratelimit": structpb.NewStructValue(&structpb.Struct{
						Fields: map[string]*structpb.Value{
							"@type":    structpb.NewStringValue("type.googleapis.com/udpa.type.v1.TypedStruct"),
							"type_url": structpb.NewStringValue("type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit"),
							"value": structpb.NewStructValue(&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"stat_prefix":  structpb.NewStringValue("http_local_rate_limiter"),
									"descriptors":  structpb.NewListValue(descriptors),
									"token_bucket": structpb.NewStructValue(getProtoStructFromProtoMessage(globalBucket)),
									"filter_enabled": structpb.NewStructValue(&structpb.Struct{
										Fields: map[string]*structpb.Value{
											"runtime_key": structpb.NewStringValue("local_rate_limit_enabled"),
											"default_value": structpb.NewStructValue(&structpb.Struct{
												Fields: map[string]*structpb.Value{
													"numerator":   structpb.NewNumberValue(100),
													"denominator": structpb.NewStringValue("HUNDRED"),
												},
											}),
										},
									}),
									"filter_enforced": structpb.NewStructValue(&structpb.Struct{
										Fields: map[string]*structpb.Value{
											"runtime_key": structpb.NewStringValue("local_rate_limit_enforced"),
											"default_value": structpb.NewStructValue(&structpb.Struct{
												Fields: map[string]*structpb.Value{
													"numerator":   structpb.NewNumberValue(100),
													"denominator": structpb.NewStringValue("HUNDRED"),
												},
											}),
										},
									}),
								},
							}),
						},
					}),
				},
			}),
		}},
	}
}

func getDescriptorConfig(tcgName, associatedApp string, quota *admiralv1.Quota, timePeriod time.Duration) *localratelimit.LocalRateLimitDescriptor {
	quotaEntry := getQuotaDescriptorEntry(tcgName, quota.Name)
	appEntry := getAssociatedAppDescriptorEntry(tcgName, quota.Name, associatedApp)

	entries := []*localratelimit.RateLimitDescriptor_Entry{quotaEntry}
	if appEntry != nil {
		entries = append(entries, appEntry)
	}

	return &localratelimit.LocalRateLimitDescriptor{
		TokenBucket: getTokenBucket(quota.MaxAmount, quota.MaxAmount, timePeriod),
		Entries:     entries,
	}
}

func getQuotaDescriptorEntry(tcgName, quotaName string) *localratelimit.RateLimitDescriptor_Entry {
	descriptorKey := getDescriptorKey(tcgName, quotaName)
	return &localratelimit.RateLimitDescriptor_Entry{Key: descriptorKey, Value: descriptorKey}
}

func getAssociatedAppDescriptorEntry(tcgName, quotaName, associatedApp string) *localratelimit.RateLimitDescriptor_Entry {
	if associatedApp == "" {
		return nil
	}

	descriptorKey := getDescriptorKey(tcgName, quotaName, associatedApp)
	return &localratelimit.RateLimitDescriptor_Entry{Key: descriptorKey, Value: descriptorKey}
}

func getRateLimitConfig(tcgName, associatedApp string, quota *admiralv1.Quota) *routev3.RateLimit {
	quotaAction := getQuotaAction(tcgName, quota)
	appAction := getAssociatedAppAction(tcgName, quota.Name, associatedApp)

	actions := []*routev3.RateLimit_Action{quotaAction}
	if appAction != nil {
		actions = append(actions, appAction)
	}
	return &routev3.RateLimit{Actions: actions}
}

func getQuotaAction(tcgName string, quota *admiralv1.Quota) *routev3.RateLimit_Action {
	headerMatchers := []*routev3.HeaderMatcher{}

	pathMatcher := getHeaderMatcher(":path", getPathRegex(quota.Path), types.REGEX)
	methodMatcher := getHeaderMatcher(":method", getMethodRegex(quota.Methods), types.REGEX)

	headerMatchers = append(headerMatchers, pathMatcher)
	headerMatchers = append(headerMatchers, methodMatcher)

	for _, header := range quota.Headers {
		headerMatcher := getHeaderMatcher(header.Name, header.Value, types.MatchCondition(header.Condition))
		headerMatchers = append(headerMatchers, headerMatcher)
	}

	descriptorKey := getDescriptorKey(tcgName, quota.Name)

	return &routev3.RateLimit_Action{
		ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
			HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
				DescriptorKey:   descriptorKey,
				DescriptorValue: descriptorKey,
				Headers:         headerMatchers,
			},
		},
	}
}

func getAssociatedAppAction(tcgName, quotaName, associatedApp string) *routev3.RateLimit_Action {
	if associatedApp == "" {
		return nil
	}

	return &routev3.RateLimit_Action{
		ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
			HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
				DescriptorKey:   getDescriptorKey(tcgName, quotaName, associatedApp),
				DescriptorValue: getDescriptorKey(tcgName, quotaName, associatedApp),
				Headers: []*routev3.HeaderMatcher{
					getHeaderMatcher(options.Params.TrafficConfigIdentityKey, associatedApp, types.EXACT),
				},
			},
		},
	}
}
