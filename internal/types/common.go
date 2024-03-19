package types

import (
	"fmt"
	"strings"
)

type FeatureName string

const (
	NaavikName             = "naavik"
	NamespaceIstioSystem   = "istio-system"
	AssetAliasMetaDataKey  = "assetAlias"
	SourceIdentityKey      = "sourceIdentity"
	DestinationIdentityKey = "destinationIdentity"

	StateCheckerNone     = "none"
	ConfigResolverSecret = "secret"

	IncludeInboundPortsAnnotation = "admiral.io/inboundPorts"

	// TrafficConfig related constants.
	TransactionIDKey        = "transactionID"
	RevisionNumberKey       = "revisionNumber"
	CreatedByKey            = "createdBy"
	CreatedTypeKey          = "createdType"
	CreatedForKey           = "createdFor"
	CreatedForEnvKey        = "createdForEnv"
	CreatedForTrafficEnvKey = "createdForTrafficEnv"
	EnvKey                  = "env"
	IsDisabledKey           = "isDisabled"
	IsTrue                  = "true"
	IsFalse                 = "false"

	AppEnvKey = "APP_ENV"

	EnvDev     = "dev"
	EnvDefault = "default"

	// Feature Names.
	FeatureVirtualService FeatureName = "virtualservice"
	FeatureRouterFilter   FeatureName = "routerfilter"
	FeatureThrottleFilter FeatureName = "throttlefilter"
	FeatureGWProxyFilter  FeatureName = "gwproxyfilter"

	// Rollout/Deployment labels.
	AppLabelKey     = "app"
	EnvLabelKey     = "env"
	ExpressLabelKey = "express"
)

type EventType string

const (
	Add    EventType = "Add"
	Update EventType = "Update"
	Delete EventType = "Delete"
)

func (et EventType) String() string {
	return string(et)
}

func (f FeatureName) String() string {
	return string(f)
}

func GetHost(env, asset, suffix string) string {
	host := fmt.Sprintf("%s.%s.%s", env, asset, suffix)
	return strings.ToLower(host)
}

type MatchCondition string

const (
	EXACT    MatchCondition = "exact"
	PREFIX   MatchCondition = "prefix"
	SUFFIX   MatchCondition = "suffix"
	CONTAINS MatchCondition = "contains"
	REGEX    MatchCondition = "regex"
)
