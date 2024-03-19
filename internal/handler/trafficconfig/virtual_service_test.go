package trafficconfig

import (
	goctx "context"
	"embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/controller"
	"github.com/intuit/naavik/internal/fake/builder"
	resourcebuilder "github.com/intuit/naavik/internal/fake/builder/resource"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/pkg/logger"
	"github.com/intuit/naavik/pkg/utils"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:embed testdata/*.json
var testdata embed.FS

//go:embed testdata/expected/*.json
var expectedTestdata embed.FS

func TestVirtualServiceGeneration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "virtualservice_test")
}

var _ = Describe("Test VirtualService Generation", func() {
	var trafficconfig *admiralv1.TrafficConfig
	var ctx context.Context
	var logMessages []string

	BeforeEach(func() {
		cache.ResetAllCaches()
		options.InitializeNaavikArgs(nil)
		ctx = context.NewContextWithLogger()
		logMessages = []string{}
		ctx.Log = ctx.Log.Hook(func(level, msg string) {
			logMessages = append(logMessages, msg)
		})
		trafficconfig = resourcebuilder.GetFakeTrafficConfig("destination-asset", "qa", "1", "test-namespace")
	})

	Context("VirtualService generation negetive cases", func() {
		When("traffic config env is empty", func() {
			It("should skip vs handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				// Remove env from traffic config
				trafficconfig.Labels = map[string]string{"test": "test"}
				trafficconfig.Annotations = map[string]string{"test": "test"}
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				Expect(logMessages).To(ContainElement("no env present in traffic config, skipping"))
			})
		})

		When("asset doesn't have dependent services", func() {
			It("should skip vs handling", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				Expect(logMessages).To(ContainElement("no dependent services found."))
			})
		})

		When("no clusters found for dependent asset", func() {
			It("should skip vs handling", func() {
				options.InitializeNaavikArgs(&options.NaavikArgs{
					IgnoreAssetAliases: []string{"ignore-asset"},
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "gw-asset")
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "ignore-asset")
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "source-asset-1")
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				Expect(logMessages).To(ContainElement("no dependent clusters found"))
			})
		})

		When("remote cluster is not in allowed scope", func() {
			It("should skip processing vs", func() {
				options.InitializeNaavikArgs(&options.NaavikArgs{
					AllowedClusterScope: []string{"allowed-cluster"},
				})
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-cluster-1")
				rc := builder.BuildRemoteCluster("ignored-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				Expect(logMessages).To(ContainElement("cluster not in allowed scope, skipping."))
			})
		})
	})

	Context("VirtualService generation positive cases", func() {
		BeforeEach(func() {
			options.InitializeNaavikArgs(&options.NaavikArgs{
				IgnoreAssetAliases:  []string{"ignore-asset"},
				AllowedClusterScope: []string{"allowed"},
			})
			// Add ignored asset and GW asset as dependent
			cache.IdentityDependency.AddDependentToIdentity("destination-asset", "gw-asset")
			cache.IdentityDependency.AddDependentToIdentity("destination-asset", "ignore-asset")
		})

		When("required resources are satisfied", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				vs, err := rc.IstioClient().GetVirtualService(ctx, "qal.destination-asset.mesh-vs", options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(err).To(BeNil())
				Expect(vs).ToNot(BeNil())
			})
		})

		When("traffic config is disabled the vs should be deleted", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("destination-asset", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				vs, err := rc.IstioClient().GetVirtualService(ctx, "qal.destination-asset.mesh-vs", options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(err).To(BeNil())
				Expect(vs).ToNot(BeNil())

				// If traffic config is disabled then vs should be deleted
				trafficconfig.Annotations["isDisabled"] = "true"
				HandleVirtualServiceForTrafficConfig(ctx, trafficconfig, statusChan)
				vs, err = rc.IstioClient().GetVirtualService(ctx, "qal.destination-asset.mesh-vs", options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).To(BeNil())
				Expect(err).ToNot(BeNil())
			})
		})

		When("Traffic Dialing without AIR env where GW is not onboarded no route config", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_without_air_trafficdial.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("Traffic Dialing without AIR env where GW is not onboarded with route config", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_without_air_trafficdial_with_routeconfig.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("Traffic dialing bdd configuration", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_dial_bdd_config.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("Traffic dialing route with a single target", func() {
			It("should generate virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic-dial-route-without-targetgroup.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("sourceIdentity is present in the context", func() {
			It("should generate virtual service only on the sourceIdentity cluster", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-2")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-2", "allowed-source-cluster-2")
				rc1 := builder.BuildRemoteCluster("allowed-source-cluster-1")
				rc2 := builder.BuildRemoteCluster("allowed-source-cluster-2")
				cache.RemoteCluster.AddCluster(rc1)
				cache.RemoteCluster.AddCluster(rc2)
				tc := getPreloadedTrafficConfig("testdata/traffic-dial-route-without-targetgroup.json")
				// Set sourceIdentity in the context
				ctx.Context = goctx.WithValue(ctx.Context, types.SourceIdentityKey, "source-asset-1")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				vs, err := rc1.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
				// Check if virtual service is not created on other source cluster
				vs, err = rc2.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(err).ToNot(BeNil())
				Expect(vs).To(BeNil())
			})
		})

		When("Virtual service on non-gw cluster and a cluster not in allowed scope", func() {
			It("should generate virtual service on only allowed cluster", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				ignoredrc := builder.BuildRemoteCluster("ignored-source-cluster-1")
				cache.RemoteCluster.AddCluster(ignoredrc)
				tc := getPreloadedTrafficConfig("testdata/traffic_dial_bdd_with_air.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())
				// Check if virtual service is not created on ignored cluster
				vs, err := ignoredrc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(err).ToNot(BeNil())
				Expect(vs).To(BeNil())

				// Check if virtual service is created on allowed cluster
				vs, err = rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("traffic config doesnot contain any routes", func() {
			It("should fail to create virtual service", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_without_any_routes.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(err).ToNot(BeNil())
				Expect(vs).To(BeNil())
				Expect(logMessages).To(ContainElement("error constructing virtual service."))
			})
		})

		When("traffic config doesnot contain any targets", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_without_any_targets.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("traffic config air", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_air.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("traffic config with custom air env", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_custom_air.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("traffic dial with same swimlane across routes", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_same_swimlane.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("App dial with different routes different dial", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic-dial-air-appdial-two-route-two-diff-swimlanes.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("App dial with two different services dial two same routes", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_two_same_servicedialroute_two_same_appdialroutes.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})

		When("App dial with same routes services dial two different routes", func() {
			It("should created desired no of routes and spec should match", func() {
				statusChan := make(chan controller.EventProcessStatus, 1)
				cache.IdentityDependency.AddDependentToIdentity("httpbin", "source-asset-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "allowed-source-cluster-1")
				cache.IdentityCluster.AddClusterToIdentity("source-asset-1", "ignored-source-cluster-1")
				rc := builder.BuildRemoteCluster("allowed-source-cluster-1")
				cache.RemoteCluster.AddCluster(rc)
				tc := getPreloadedTrafficConfig("testdata/traffic_config_two_different_servicedial_routes_two_different_appdialroute.json")
				HandleVirtualServiceForTrafficConfig(ctx, tc, statusChan)

				vsName := getVirtualServiceName(utils.TrafficConfigUtil(tc).GetEnv(), utils.TrafficConfigUtil(tc).GetIdentityLowerCase())

				vs, err := rc.IstioClient().GetVirtualService(ctx, vsName, options.GetSyncNamespace(), metav1.GetOptions{})
				Expect(vs).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})
})

func getNoOfAppDialsForRoute(route *admiralv1.Route, edgeService *admiralv1.EdgeService) int {
	var noOFAppDials int
	if route.Config != nil {
		for _, config := range route.Config {
			for _, tg := range edgeService.TargetGroups {
				for _, tgAppOverride := range tg.AppOverrides {
					for _, wg := range tgAppOverride.Weights {
						if wg.Name == config.TargetGroupSelector {
							noOFAppDials += 1
							break
						}
					}
				}
			}
			break
		}
	}
	return noOFAppDials
}

func getNoOfServiceDialsForRoute(route *admiralv1.Route, edgeService *admiralv1.EdgeService) int {
	var noOFServiceDials int
	if route.Config != nil {
		serviceRoute, ifYes := checkIfRouteDoesNotHaveTargetGroupAssociated(route.Config)
		if ifYes {
			return serviceRoute
		}
		for _, tg := range edgeService.TargetGroups {
			for _, config := range route.Config {
				for _, wg := range tg.Weights {
					if wg.Name == config.TargetGroupSelector {
						noOFServiceDials += 1
						break
					}
				}
				break
			}
		}
	}
	return noOFServiceDials
}

func checkIfRouteDoesNotHaveTargetGroupAssociated(config []*admiralv1.Config) (int, bool) {
	var noOFServiceDials int
	if len(config) == 1 {
		for _, tg := range config {
			if tg.TargetGroupSelector == "" {
				noOFServiceDials = 1
				return noOFServiceDials, true
			}
		}
	}
	return 0, false
}

func getNoOFDefaultRoutes(tc *admiralv1.TrafficConfig) int {
	var airEnv int
	for _, env := range tc.Spec.WorkloadEnv {
		if strings.Contains(env, "-air") {
			airEnv += 1
		}
	}
	return len(tc.Spec.WorkloadEnv) + airEnv
}

func getExpectedNoOfVSHTTPRoutes(tc *admiralv1.TrafficConfig) int {
	var noOFRoutes int
	if tc.Annotations["isDisabled"] == "true" {
		return 0
	}
	var noOFAppDials, noOFServiceDials int
	for _, route := range tc.Spec.EdgeService.Routes {
		if route.Config != nil {
			x := getNoOfAppDialsForRoute(route, tc.Spec.EdgeService)
			noOFAppDials = noOFAppDials + x
			y := getNoOfServiceDialsForRoute(route, tc.Spec.EdgeService)
			noOFServiceDials = noOFServiceDials + y
			noOFRoutes = noOFAppDials + noOFServiceDials
		} else {
			// noOFRoutes = len(tc.Spec.EdgeService.Routes)
			noOFRoutes += len(route.WorkloadEnvSelectors)
		}
	}
	noOFDefaultRoutes := getNoOFDefaultRoutes(tc)
	noOFRoutes = noOFRoutes + noOFDefaultRoutes
	return noOFRoutes
}

func getPreloadedTrafficConfig(filename string) *admiralv1.TrafficConfig {
	tc := &admiralv1.TrafficConfig{}
	tc_data, err := testdata.ReadFile(filename)
	if err != nil {
		logger.Log.Infof("Error reading %s: %v", filename, err)
	}
	err = json.Unmarshal(tc_data, tc)
	if err != nil {
		logger.Log.Fatalf("failed to unmarshal traffic config : %+v", tc_data, err)
	}
	return tc
}
