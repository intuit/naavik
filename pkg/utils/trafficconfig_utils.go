package utils

import (
	"strings"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	admiralv1 "github.com/istio-ecosystem/admiral-api/pkg/apis/admiral/v1"
)

type TrafficConfigInterface interface {
	GetIdentity() string
	GetIdentityLowerCase() string
	GetRevision() string
	GetTransactionID() string
	GetEnv() string
	GetWorkloadEnvs() []string
	GetEdgeService() *admiralv1.EdgeService
	GetQuotaGroup() *admiralv1.QuotaGroup
	IsIgnored() bool
	IsDisabled() bool
	GetTrafficConfig() *admiralv1.TrafficConfig
}

type trafficConfigUtil struct {
	tc *admiralv1.TrafficConfig
}

func TrafficConfigUtil(tc *admiralv1.TrafficConfig) TrafficConfigInterface {
	return &trafficConfigUtil{
		tc: tc,
	}
}

func (tcu *trafficConfigUtil) GetIdentity() string {
	identity := ""
	if tcu.tc == nil {
		return ""
	}
	if tcu.tc.Labels != nil {
		identity = tcu.tc.Labels[options.GetTrafficConfigIdentityKey()]
	}
	if len(identity) == 0 && tcu.tc.Annotations != nil {
		identity = tcu.tc.Annotations[options.GetTrafficConfigIdentityKey()]
	}
	return identity
}

func (tcu *trafficConfigUtil) GetIdentityLowerCase() string {
	return strings.ToLower(tcu.GetIdentity())
}

func (tcu *trafficConfigUtil) GetRevision() string {
	revision := ""
	if tcu.tc == nil {
		return ""
	}
	if tcu.tc.Labels != nil {
		revision = tcu.tc.Labels[types.RevisionNumberKey]
	}
	if len(revision) == 0 && tcu.tc.Annotations != nil {
		revision = tcu.tc.Annotations[types.RevisionNumberKey]
	}
	return revision
}

func (tcu *trafficConfigUtil) GetTransactionID() string {
	tid := ""
	if tcu.tc == nil {
		return tid
	}
	if tcu.tc.Labels != nil {
		tid = tcu.tc.Labels[types.TransactionIDKey]
	}
	if len(tid) == 0 && tcu.tc.Annotations != nil {
		tid = tcu.tc.Annotations[types.TransactionIDKey]
	}
	return tid
}

func (tcu *trafficConfigUtil) GetEnv() string {
	env := ""
	if tcu.tc == nil {
		return ""
	}
	if tcu.tc.Labels != nil {
		env = tcu.tc.Labels[types.EnvKey]
	}
	if len(env) == 0 && tcu.tc.Annotations != nil {
		env = tcu.tc.Annotations[types.EnvKey]
	}
	return env
}

func (tcu *trafficConfigUtil) IsIgnored() bool {
	if tcu.tc == nil {
		return false
	}
	if tcu.tc.Labels != nil && len(tcu.tc.Labels[options.GetResourceIgnoreLabel()]) > 0 {
		return tcu.tc.Labels[options.GetResourceIgnoreLabel()] == types.IsTrue
	}
	if tcu.tc.Annotations != nil && len(tcu.tc.Annotations[options.GetResourceIgnoreLabel()]) > 0 {
		return tcu.tc.Annotations[options.GetResourceIgnoreLabel()] == types.IsTrue
	}
	return false
}

func (tcu *trafficConfigUtil) IsDisabled() bool {
	if tcu.tc == nil {
		return false
	}
	if tcu.tc.Labels != nil && len(tcu.tc.Labels[types.IsDisabledKey]) > 0 {
		return tcu.tc.Labels[types.IsDisabledKey] == types.IsTrue
	}
	if tcu.tc.Annotations != nil && len(tcu.tc.Annotations[types.IsDisabledKey]) > 0 {
		return tcu.tc.Annotations[types.IsDisabledKey] == types.IsTrue
	}
	return false
}

func (tcu *trafficConfigUtil) GetEdgeService() *admiralv1.EdgeService {
	return tcu.tc.Spec.EdgeService
}

func (tcu *trafficConfigUtil) GetQuotaGroup() *admiralv1.QuotaGroup {
	return tcu.tc.Spec.QuotaGroup
}

func (tcu *trafficConfigUtil) GetTrafficConfig() *admiralv1.TrafficConfig {
	return tcu.tc
}

func (tcu *trafficConfigUtil) GetWorkloadEnvs() []string {
	if tcu.tc == nil {
		return []string{}
	}
	return tcu.tc.Spec.WorkloadEnv
}
