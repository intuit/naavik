package trafficconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/internal/types/remotecluster"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppDialingDetails map[string]map[string]map[string]int // map[TGgroupName][AppAssetName][hostName][weightPercentage]  == per targetGroup - group all app assets and then host Details

type RouteDetails struct {
	RouteName          string
	Inbound            string
	Outbound           string
	ConfigDetails      []*routeTargetInfo
	Timeout            *durationpb.Duration
	ServiceDialDetails []*endpointWeight
	AppDialingDetails  map[string][]*endpointWeight
}
type endpointWeight struct {
	endpoint string
	weight   int
}

type routeTargetInfo struct {
	targetGroupSelector string
	meshEndpoint        string
}

func HandleVirtualServiceForTrafficConfig(ctx context.Context, trafficconfig *admiralv1.TrafficConfig, _ chan controller.EventProcessStatus) {
	tcUtil := utils.TrafficConfigUtil(trafficconfig)
	if len(tcUtil.GetEnv()) == 0 {
		ctx.Log.Error("no env present in traffic config, skipping")
		return
	}

	dependents := cache.IdentityDependency.GetDependentsForIdentity(tcUtil.GetIdentity())
	if len(dependents) == 0 {
		ctx.Log.Str(logger.WorkloadIdentifierKey, tcUtil.GetIdentity()).Infof("no dependent services found.")
		return
	}

	handleVirtualServiceForMeshDependents(ctx, dependents, tcUtil)
}

func handleVirtualServiceForMeshDependents(ctx context.Context, dependents []string, tc utils.TrafficConfigInterface) {
	dependentClusters := make(map[string]string)
	for _, depIdentity := range dependents {
		isIgnoredAsset := options.IsAssetIgnored(depIdentity)
		if options.IsAssetIgnored(depIdentity) {
			ctx.Log.Str(logger.WorkloadIdentifierKey, depIdentity).Bool("isIgnoredAsset", isIgnoredAsset).Infof("ignoring this dependent identity")
			continue
		}
		// Check if the event should be handled only for a sourceIdentity
		sourceIdentityCtx := ctx.Context.Value(types.SourceIdentityKey)
		var sourceIdentity string
		if sourceIdentityCtx != nil {
			sourceIdentity = sourceIdentityCtx.(string)
		}
		if len(sourceIdentity) > 0 && sourceIdentity != depIdentity {
			ctx.Log.Str(logger.DependentIdentityKey, depIdentity).Str(logger.SourceAssetKey, sourceIdentity).Info("Handling only for dependent is equal to sourceIdentity, skipping.")
			continue
		}

		clusters := cache.IdentityCluster.GetClustersForIdentity(depIdentity)
		for _, cluster := range clusters {
			dependentClusters[cluster] = cluster
		}
	}
	if len(dependentClusters) == 0 {
		ctx.Log.Warn("no dependent clusters found")
		return
	}
	ctx.Log.Str(logger.ResourceIdentifierKey, tc.GetIdentity()).Any("dependentClusters", dependentClusters).Info("dependent clusters")
	vsMap, err := buildVirtualServiceForMeshDependents(ctx, tc)
	if err != nil {
		ctx.Log.Str(logger.ErrorKey, err.Error()).Error("error constructing virtual service.")
		return
	}
	for clusterID := range dependentClusters {
		if !options.IsClusterInAllowedScope(clusterID) {
			ctx.Log.Str(logger.HandlerNameKey, "VirtualService").Str(logger.ClusterKey, clusterID).Warnf("cluster not in allowed scope, skipping.")
			continue
		}
		rc, found := cache.RemoteCluster.GetCluster(clusterID)
		if !found {
			ctx.Log.Str(logger.ClusterKey, clusterID).Info("remote cluster not found, skipping.")
			continue
		}
		createUpdateDeleteVirtualServices(ctx, rc, vsMap, tc)
	}
}

// Form match rules here
// we create virtual service per workloadEnv , but scan through all routes in edgeSpec ,
// we will just keep constructing virtual service by going through all routes and see if that route is present in that env, If yes then we will add that route in that VS's route.
// loop through config array per route
//
// construct match cases per route.
// construct app dial asset details like assetAlias and weights per route.
func getMeshDNSForTargetGroup(tc utils.TrafficConfigInterface, configDetails *admiralv1.Config, route *admiralv1.Route) string {
	if tc.GetEdgeService().Targets == nil {
		return ""
	}

	if len(route.WorkloadEnvSelectors) == 0 {
		return ""
	}

	finalTargetWorkloadENV := route.WorkloadEnvSelectors[0]

	for _, target := range tc.GetTrafficConfig().Spec.EdgeService.Targets {
		if target.Name == configDetails.TargetSelector {
			if len(target.MeshDNS) == 0 {
				return types.GetHost(finalTargetWorkloadENV, tc.GetIdentityLowerCase(), options.GetHostnameSuffix())
			}
			return strings.ToLower(target.MeshDNS)
		}
	}

	return finalTargetWorkloadENV
}

func getCombinedRouteDetails(tc utils.TrafficConfigInterface, AppDialingAcrossAllRoutes AppDialingDetails, serviceDialDetails map[string]int) []*RouteDetails {
	allRouteDetail := make([]*RouteDetails, 0)
	for _, route := range tc.GetTrafficConfig().Spec.EdgeService.Routes {
		var routeDetails RouteDetails
		routeDetails.RouteName = route.Name
		routeDetails.Inbound = route.Inbound
		routeDetails.Outbound = route.Outbound
		routeDetails.Timeout = getRouteTimeout(route)
		routeConfigs := make([]*routeTargetInfo, 0)
		if route.Config == nil {
			allRouteDetail = append(allRouteDetail, &routeDetails)
			continue
		}
		for _, configDetails := range route.Config {
			meshDNS := getMeshDNSForTargetGroup(tc, configDetails, route)
			rInfo := &routeTargetInfo{
				targetGroupSelector: configDetails.TargetGroupSelector,
				meshEndpoint:        meshDNS,
			}
			routeConfigs = append(routeConfigs, rInfo)
			routeDetails.ConfigDetails = routeConfigs
		}
		routeAppDialingDetails := make(map[string][]*endpointWeight)
		constructAppDialRoutes(AppDialingAcrossAllRoutes[route.Name], routeDetails, routeAppDialingDetails)
		routeDetails.ServiceDialDetails = constructServiceDialRoute(serviceDialDetails, routeConfigs)
		routeDetails.AppDialingDetails = routeAppDialingDetails
		allRouteDetail = append(allRouteDetail, &routeDetails)
	}
	return allRouteDetail
}

func constructServiceDialRoute(serviceDialDetails map[string]int, routeTargetInfos []*routeTargetInfo) []*endpointWeight {
	serviceWeightMap := make([]*endpointWeight, 0)
	var routeInfo *routeTargetInfo

	if len(routeTargetInfos) == 1 && routeTargetInfos[0].targetGroupSelector == "" {
		endpoint := &endpointWeight{
			endpoint: routeTargetInfos[0].meshEndpoint,
			weight:   100,
		}
		serviceWeightMap = append(serviceWeightMap, endpoint)
		return serviceWeightMap
	}
	for _, routeInfo = range routeTargetInfos {
		if weight, ok := serviceDialDetails[routeInfo.targetGroupSelector]; ok {
			endpoint := &endpointWeight{
				endpoint: routeInfo.meshEndpoint,
				weight:   weight,
			}
			serviceWeightMap = append(serviceWeightMap, endpoint)
		}
	}
	return serviceWeightMap
}

func constructAppDialRoutes(AppDialingAcrossAllRoutes map[string]map[string]int, routeDetails RouteDetails, routeAppDialingDetails map[string][]*endpointWeight) {
	for appAssetAlias, appWeights := range AppDialingAcrossAllRoutes {
		appWeightMap := make([]*endpointWeight, 0)
		for name, weight := range appWeights {
			for _, routeConfig := range routeDetails.ConfigDetails {
				if routeConfig.targetGroupSelector == name {
					endpoint := &endpointWeight{
						endpoint: routeConfig.meshEndpoint,
						weight:   weight,
					}
					appWeightMap = append(appWeightMap, endpoint)
				}
			}
		}
		if len(appWeightMap) != 0 {
			routeAppDialingDetails[appAssetAlias] = appWeightMap
		}
	}
}

func buildMatchRulesForVirtualService(ctx context.Context, allRouteDetails []*RouteDetails, tc utils.TrafficConfigInterface) []*networkingv1alpha3.HTTPRoute {
	allRulesPerRoute := make([]*networkingv1alpha3.HTTPRoute, 0)
	for _, route := range allRouteDetails {
		// check if there are app dials present in the routeDetails, If yes then create match rules for them
		if len(route.AppDialingDetails) != 0 {
			appDialRule := buildAppDialRules(ctx, route)
			allRulesPerRoute = append(allRulesPerRoute, appDialRule...)
		}
		if len(route.ServiceDialDetails) != 0 {
			serviceDialRule := buildServiceDialRule(ctx, route)
			allRulesPerRoute = append(allRulesPerRoute, serviceDialRule)
		}
		// this will happen when the route does not have config for a route.
		if len(route.ServiceDialDetails) == 0 && len(route.AppDialingDetails) == 0 {
			httpRule := buildRuleWhenRouteConfigEmpty(ctx, route, tc)
			allRulesPerRoute = append(allRulesPerRoute, httpRule...)
		}
	}
	return allRulesPerRoute
}

func buildAppDialRules(ctx context.Context, route *RouteDetails) []*networkingv1alpha3.HTTPRoute {
	httpRoutes := make([]*networkingv1alpha3.HTTPRoute, 0)
	for appAssetAlias, routeWeights := range route.AppDialingDetails {
		routes := make([]*networkingv1alpha3.HTTPRouteDestination, 0)
		for _, endpoint := range routeWeights {
			routeDestination := networkingv1alpha3.HTTPRouteDestination{
				Destination: &networkingv1alpha3.Destination{Host: endpoint.endpoint},
				Weight:      int32(endpoint.weight),
			}
			routes = append(routes, &routeDestination)
		}

		httpRoute := &networkingv1alpha3.HTTPRoute{
			Name: route.RouteName + "-" + appAssetAlias,
			Match: []*networkingv1alpha3.HTTPMatchRequest{
				{
					Uri: &networkingv1alpha3.StringMatch{
						MatchType: &networkingv1alpha3.StringMatch_Prefix{Prefix: route.Inbound},
					},
				},
			},
			Route:   routes,
			Rewrite: &networkingv1alpha3.HTTPRewrite{Uri: route.Outbound},
			Timeout: getFinalRouteTimeout(route.Timeout),
		}

		ctx.Log.Str(logger.RouteNameKey, httpRoute.String()).Debug("adding app dial route.")
		httpRoutes = append(httpRoutes, httpRoute)
	}
	return httpRoutes
}

func getFinalRouteTimeout(timeout *durationpb.Duration) *durationpb.Duration {
	if timeout.AsDuration().Nanoseconds() > time.Millisecond.Nanoseconds() {
		return timeout
	}
	return nil
}

func buildServiceDialRule(ctx context.Context, route *RouteDetails) *networkingv1alpha3.HTTPRoute {
	serviceHTTPRoute := &networkingv1alpha3.HTTPRoute{
		Name: route.RouteName,
		Match: []*networkingv1alpha3.HTTPMatchRequest{
			{
				Uri: &networkingv1alpha3.StringMatch{
					MatchType: &networkingv1alpha3.StringMatch_Prefix{
						Prefix: route.Inbound,
					},
				},
			},
		},
		Rewrite: &networkingv1alpha3.HTTPRewrite{Uri: route.Outbound},
		Timeout: getFinalRouteTimeout(route.Timeout),
	}

	for _, endpointDetails := range route.ServiceDialDetails {
		routeDestination := networkingv1alpha3.HTTPRouteDestination{}
		routeDestination.Weight = int32(endpointDetails.weight)
		routeDestination.Destination = &networkingv1alpha3.Destination{
			Host: endpointDetails.endpoint,
		}
		serviceHTTPRoute.Route = append(serviceHTTPRoute.Route, &routeDestination)
	}
	ctx.Log.Str(logger.RouteNameKey, serviceHTTPRoute.String()).Debug("adding service dial route.")
	return serviceHTTPRoute
}

func buildRuleWhenRouteConfigEmpty(ctx context.Context, routeDetails *RouteDetails, tc utils.TrafficConfigInterface) []*networkingv1alpha3.HTTPRoute {
	httpRoutes := make([]*networkingv1alpha3.HTTPRoute, 0)
	for _, route := range tc.GetTrafficConfig().Spec.EdgeService.Routes {
		if route.Name == routeDetails.RouteName {
			for _, env := range route.WorkloadEnvSelectors {
				serviceHTTPRoute := &networkingv1alpha3.HTTPRoute{
					Name: routeDetails.RouteName + "-" + env, // added route name
					Match: []*networkingv1alpha3.HTTPMatchRequest{
						{
							Uri: &networkingv1alpha3.StringMatch{
								MatchType: &networkingv1alpha3.StringMatch_Prefix{Prefix: route.Inbound},
							},
						},
					},
					Route: []*networkingv1alpha3.HTTPRouteDestination{
						{
							Destination: &networkingv1alpha3.Destination{
								Host: types.GetHost(env, tc.GetIdentityLowerCase(), options.GetHostnameSuffix()),
							},
						},
					},
					Rewrite: &networkingv1alpha3.HTTPRewrite{Uri: route.Outbound},
					Timeout: getFinalRouteTimeout(routeDetails.Timeout),
				}
				ctx.Log.Str(logger.RouteNameKey, serviceHTTPRoute.String()).Debug("adding empty config route.")
				httpRoutes = append(httpRoutes, serviceHTTPRoute)
			}
		}
	}

	return httpRoutes
}

func addAllowAllRule(ctx context.Context, tc utils.TrafficConfigInterface, allRulesPerRoute []*networkingv1alpha3.HTTPRoute) []*networkingv1alpha3.HTTPRoute {
	var timeout *durationpb.Duration
	for _, httpRoute := range allRulesPerRoute {
		if httpRoute.Match != nil {
			if httpRoute.Timeout.GetSeconds() > timeout.GetSeconds() {
				timeout = httpRoute.Timeout
			}
		}
	}
	for _, env := range tc.GetTrafficConfig().Spec.WorkloadEnv {
		assetAlias := tc.GetIdentityLowerCase()
		authorityHost := types.GetHost(env, assetAlias, options.GetHostnameSuffix())
		allowAllRoute := createAllowRouteMatch(env, timeout, authorityHost, "defaultall"+"-"+env, assetAlias)
		ctx.Log.Str(logger.RouteNameKey, allowAllRoute.String()).Debug("adding allow all route.")
		allRulesPerRoute = append(allRulesPerRoute, allowAllRoute)
	}
	return allRulesPerRoute
}

func createAllowRouteMatch(env string, timeout *durationpb.Duration, authorityHost string, routeName string, assetAlias string) *networkingv1alpha3.HTTPRoute {
	allowAllRoute := &networkingv1alpha3.HTTPRoute{}
	allowAllRoute.Name = routeName
	finaTimeout := getFinalRouteTimeout(timeout)
	if finaTimeout != nil {
		allowAllRoute.Timeout = timeout
	}
	allowAllMatchRequest := &networkingv1alpha3.HTTPMatchRequest{}
	allowAllMatchRequest.Authority = &networkingv1alpha3.StringMatch{
		MatchType: &networkingv1alpha3.StringMatch_Prefix{Prefix: strings.ToLower(authorityHost)},
	}
	allowAllMatchRequest.Uri = &networkingv1alpha3.StringMatch{
		MatchType: &networkingv1alpha3.StringMatch_Prefix{Prefix: "/"},
	}

	allowAllDefRoute := &networkingv1alpha3.HTTPRouteDestination{}
	allowAllDefRoute.Destination = &networkingv1alpha3.Destination{
		Host: types.GetHost(env, assetAlias, options.GetHostnameSuffix()),
	}
	allowAllRoute.Route = append(allowAllRoute.Route, allowAllDefRoute)
	allowAllRoute.Match = append(allowAllRoute.Match, allowAllMatchRequest)
	return allowAllRoute
}

// buildVirtualServiceWithoutTargetDetails constructs virtual service rules - when a service does not have any target OR target groups present.
func buildVirtualServiceWithoutTargetDetails(ctx context.Context, tc utils.TrafficConfigInterface) *v1alpha3.VirtualService {
	if tc.GetTrafficConfig().Spec.EdgeService != nil {
		httpRoutes := make([]*networkingv1alpha3.HTTPRoute, 0)
		for _, route := range tc.GetTrafficConfig().Spec.EdgeService.Routes {
			httpRoutesPerRoute := constructVSRouteWithoutTargetDetails(ctx, tc, route)
			httpRoutes = append(httpRoutes, httpRoutesPerRoute...)
		}
		allRules := addAllowAllRule(ctx, tc, httpRoutes)
		return buildVirtualServiceObject(tc, allRules)
	}
	return nil
}

func constructVSRouteWithoutTargetDetails(ctx context.Context, tc utils.TrafficConfigInterface, route *admiralv1.Route) []*networkingv1alpha3.HTTPRoute {
	httpRoutes := make([]*networkingv1alpha3.HTTPRoute, 0)
	if route.Config == nil {
		httpRoute := buildRuleWhenRouteConfigEmpty(ctx, &RouteDetails{
			RouteName: route.Name,
			Timeout:   durationpb.New(time.Duration(route.Timeout)),
		}, tc)
		httpRoutes = append(httpRoutes, httpRoute...)
	}
	return httpRoutes
}

func getRouteTimeout(route *admiralv1.Route) *durationpb.Duration {
	return &durationpb.Duration{
		Seconds: int64(route.Timeout / 1000),
		Nanos:   int32(route.Timeout%1000) * 1000 * 1000,
	}
}

// TODO: cognitive score is 56, improve to make it below 20
//
//nolint:gocognit
func getTargetGroupRouteNameMap(tc utils.TrafficConfigInterface) map[string][]string {
	targetGroupRouteNameMap := make(map[string][]string) // tgGroup = routeName
	for _, route := range tc.GetEdgeService().Routes {
		if route.Config != nil {
			for _, routeConfig := range route.Config {
				for _, targetGroup := range tc.GetEdgeService().TargetGroups {
					routeList := make([]string, 0)
					for _, tgWeight := range targetGroup.Weights {
						if tgWeight.Name == routeConfig.TargetGroupSelector {
							if v, ok := targetGroupRouteNameMap[targetGroup.Name]; ok {
								v = append(v, route.Name)
								targetGroupRouteNameMap[targetGroup.Name] = v
							} else {
								routeList = append(routeList, route.Name)
								targetGroupRouteNameMap[targetGroup.Name] = routeList
							}
						}
					}
					for _, appOverride := range targetGroup.AppOverrides {
						for _, appWeight := range appOverride.Weights {
							if appWeight.Name == routeConfig.TargetGroupSelector {
								if v, ok := targetGroupRouteNameMap[targetGroup.Name]; ok {
									v = append(v, route.Name)
									targetGroupRouteNameMap[targetGroup.Name] = v
								} else {
									routeList = append(routeList, route.Name)
									targetGroupRouteNameMap[targetGroup.Name] = routeList
								}
							}
						}
					}
				}
			}
		}
	}
	return targetGroupRouteNameMap
}

func buildVirtualServiceForMeshDependents(ctx context.Context, tc utils.TrafficConfigInterface) (*v1alpha3.VirtualService, error) {
	vs := &v1alpha3.VirtualService{}
	if tc.GetEdgeService() != nil {
		if len(tc.GetEdgeService().Routes) == 0 {
			return nil, fmt.Errorf("no routes present")
		}
		if tc.GetEdgeService().Targets == nil || tc.GetEdgeService().TargetGroups == nil {
			ctx.Log.Info("building virtual service without target details.")
			vs = buildVirtualServiceWithoutTargetDetails(ctx, tc)
			if vs == nil {
				return nil, fmt.Errorf("service does not have edgeservice config details")
			}
			return vs, nil
		}
		ctx.Log.Info("building virtual service with target details.")
		tgroupRouteNameMap := getTargetGroupRouteNameMap(tc)
		appDialingPerTG := AppDialingDetails{}
		serviceDialDetails := make(map[string]int) // weights.Name across targetGroups are unique and hence we can use a common map.
		buildTargetGroupRouteNameMap(tc, tgroupRouteNameMap, appDialingPerTG, serviceDialDetails)
		routeDetails := getCombinedRouteDetails(tc, appDialingPerTG, serviceDialDetails)
		allHTTPRules := buildMatchRulesForVirtualService(ctx, routeDetails, tc)
		allRules := addAllowAllRule(ctx, tc, allHTTPRules)
		vs = buildVirtualServiceObject(tc, allRules)
	}
	return vs, nil
}

// TODO: cognitive score is 21, improve to make it below 20
//
//nolint:gocognit
func buildTargetGroupRouteNameMap(tc utils.TrafficConfigInterface, tgRouteMap map[string][]string, appDialingPerTG AppDialingDetails, serviceDialDetails map[string]int) {
	for _, tGroup := range tc.GetEdgeService().TargetGroups {
		if tGroup.AppOverrides != nil {
			appAssetToWeightPercentageMap := make(map[string]map[string]int) // sample map ->  map[httpbin:map[custom:75 express:25]]
			for _, appOverride := range tGroup.AppOverrides {
				weightMap := make(map[string]int)
				for _, i := range appOverride.Weights {
					weightMap[i.Name] = i.Weight
				}
				appAssetToWeightPercentageMap[appOverride.AssetAlias] = weightMap
				if routes, ok := tgRouteMap[tGroup.Name]; ok {
					for _, routeName := range routes {
						appDialingPerTG[routeName] = appAssetToWeightPercentageMap
					}
				}
			}
		}
		for _, weights := range tGroup.Weights {
			serviceDialDetails[weights.Name] = weights.Weight
		}
	}
}

func checkIfHostAlreadyExists(vsHosts []string, host string) bool {
	for _, vsHost := range vsHosts {
		if vsHost == host {
			return true
		}
	}
	return false
}

func constructHostsForVirtualService(envs []string, assetAlias string) []string {
	vsHosts := make([]string, 0)
	for _, env := range envs {
		meshHost := types.GetHost(env, assetAlias, options.GetHostnameSuffix())
		if !checkIfHostAlreadyExists(vsHosts, meshHost) {
			vsHosts = append(vsHosts, strings.ToLower(meshHost))
		}
	}

	return vsHosts
}

func getVirtualServiceName(env string, assetAlias string) string {
	const vsSuffix = "vs"
	var envPrefix string
	switch env {
	case "qa":
		envPrefix = "qal"
	case "e2e":
		envPrefix = "e2e"
	case "prd":
		envPrefix = "prd"
	case "prf":
		envPrefix = "prf"
	default:
		return ""
	}
	return fmt.Sprintf("%s-%s", types.GetHost(envPrefix, assetAlias, options.GetHostnameSuffix()), vsSuffix)
}

func buildVirtualServiceObject(tc utils.TrafficConfigInterface, vsRoutes []*networkingv1alpha3.HTTPRoute) *v1alpha3.VirtualService {
	vs := &v1alpha3.VirtualService{} // form virtual service object here
	vs.Name = getVirtualServiceName(tc.GetEnv(), tc.GetIdentityLowerCase())
	vs.Namespace = options.GetSyncNamespace()
	vs.Spec.Hosts = constructHostsForVirtualService(tc.GetWorkloadEnvs(), tc.GetIdentityLowerCase())
	vsAnnotations := make(map[string]string)
	vsAnnotations[types.RevisionNumberKey] = tc.GetRevision()
	vsAnnotations[types.TransactionIDKey] = tc.GetTransactionID()
	vsAnnotations[types.CreatedForEnvKey] = strings.Join(tc.GetWorkloadEnvs(), "_")
	vs.ObjectMeta.SetAnnotations(vsAnnotations)

	vsLabels := make(map[string]string)
	vsLabels[types.CreatedForKey] = strings.ToLower(tc.GetIdentity())
	vsLabels[types.CreatedByKey] = types.NaavikName
	vsLabels[types.CreatedForTrafficEnvKey] = tc.GetEnv()
	vs.ObjectMeta.SetLabels(vsLabels)
	vs.Spec.Http = make([]*networkingv1alpha3.HTTPRoute, 0)
	vs.Spec.Http = append(vs.Spec.Http, vsRoutes...)
	return vs
}

func createUpdateDeleteVirtualServices(ctx context.Context, rc remotecluster.RemoteCluster, vs *v1alpha3.VirtualService, tc utils.TrafficConfigInterface) {
	if tc.IsDisabled() {
		rc.IstioClient().DeleteVirtualService(ctx, vs.Name, options.GetSyncNamespace(), metav1.DeleteOptions{})
		return
	}
	existingVs, err := rc.IstioClient().GetVirtualService(ctx, vs.Name, options.GetSyncNamespace(), metav1.GetOptions{})

	if existingVs != nil && err == nil {
		vs.ObjectMeta.SetResourceVersion(existingVs.ResourceVersion)
		rc.IstioClient().UpdateVirtualService(ctx, vs, metav1.UpdateOptions{})
	} else {
		rc.IstioClient().CreateVirtualService(ctx, vs, metav1.CreateOptions{})
	}
}
