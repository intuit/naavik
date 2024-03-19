package trafficconfig

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/cache"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
	"github.com/intuit/naavik/internal/types/remotecluster"
	"github.com/intuit/naavik/pkg/utils"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	networkingv1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func listRateLimitingFilters(ctx context.Context, rc remotecluster.RemoteCluster, tcUtil utils.TrafficConfigInterface) (*networkingv1alpha3.EnvoyFilterList, error) {
	labelSet := metav1.LabelSelector{
		MatchLabels: map[string]string{
			types.CreatedForKey:           tcUtil.GetIdentityLowerCase(),
			types.CreatedForTrafficEnvKey: tcUtil.GetEnv(),
		},
	}

	return rc.IstioClient().ListEnvoyFilters(ctx, types.NamespaceIstioSystem, metav1.ListOptions{LabelSelector: labels.Set(labelSet.MatchLabels).String()})
}

func getWorkLoadLabels(_ context.Context, clusterID string, identity string, workLoadEnv string) (map[string]string, error) {
	labels := make(map[string]string)
	deploy := cache.Deployments.GetByClusterIdentityEnv(clusterID, identity, workLoadEnv)
	if deploy != nil {
		if val, ok := deploy.Labels["app"]; ok {
			labels["app"] = val
			labels["assetAlias"] = identity
			labels["env"] = workLoadEnv
			return labels, nil
		}
	}
	if options.IsArgoRolloutsEnabled() {
		rollout := cache.Rollouts.GetByClusterIdentityEnv(clusterID, identity, workLoadEnv)
		if rollout != nil {
			if val, ok := rollout.Labels["app"]; ok {
				labels["app"] = val
				labels["assetAlias"] = identity
				labels["env"] = workLoadEnv
				return labels, nil
			}
		}
	}
	return nil, fmt.Errorf("matching workload does not exist for identity=%s and workloadEnv=%s", identity, workLoadEnv)
}

func contains(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}

	return false
}

func getDescriptorKey(s ...string) string {
	key := strings.Join(s, ":")
	return base64.StdEncoding.EncodeToString([]byte(key))
}

func getPathRegex(path string) string {
	path = strings.ReplaceAll(path, ".*", "*")
	return strings.ReplaceAll(path, "*", ".*")
}

func getMethodRegex(methods []string) string {
	if len(methods) == 0 {
		return ".*"
	}

	m := []string{}
	for _, method := range methods {
		m = append(m, strings.ToLower(method))
		m = append(m, strings.ToUpper(method))
	}
	return strings.Join(m, "|")
}

func getHeaderMatcher(name, value string, condition types.MatchCondition) *routev3.HeaderMatcher {
	var matcher *matcherv3.StringMatcher

	switch condition {
	case types.EXACT:
		matcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: value,
			},
		}
	case types.PREFIX:
		matcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Prefix{
				Prefix: value,
			},
		}
	case types.SUFFIX:
		matcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Suffix{
				Suffix: value,
			},
		}
	case types.CONTAINS:
		matcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Contains{
				Contains: value,
			},
		}
	case types.REGEX:
		matcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_SafeRegex{
				SafeRegex: &matcherv3.RegexMatcher{
					Regex: value,
				},
			},
		}
	}

	return &routev3.HeaderMatcher{
		Name:                 name,
		HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{StringMatch: matcher},
	}
}

func getProtoStructFromProtoMessage(msg protoreflect.ProtoMessage) *structpb.Struct {
	m, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
	if err != nil {
		panic(err)
	}

	s := &structpb.Struct{}
	if err := protojson.Unmarshal(m, s); err != nil {
		panic(err)
	}

	return s
}

func getTokenBucket(maxTokens, tokensPerFill int, fillInterval time.Duration) *typev3.TokenBucket {
	return &typev3.TokenBucket{
		MaxTokens:     uint32(maxTokens),
		FillInterval:  durationpb.New(fillInterval),
		TokensPerFill: wrapperspb.UInt32(uint32(tokensPerFill)),
	}
}
