package logger

const (
	// Standard keys for structured logging.
	TimestampKey    = "ts"
	LevelKey        = "level"
	ErrorKey        = "error"
	MessageKey      = "msg"
	EventType       = "eventType"
	ChildEventIDKey = "childEventId"
	EventIDKey      = "eventId"
	EventStatusKey  = "status"
	TimeTakenMSKey  = "timeTakenMS"
	WaitTimeMSKey   = "waitTimeMS"
	QueueTimeMSKey  = "queueTimeMS"
	QueueLengthKey  = "queueLen"
	NameKey         = "name"
	OperationKey    = "op"
	TypeKey         = "type"
	DeleteValue     = "delete"
	CreateValue     = "create"
	UpdateValue     = "update"
	EnvKey          = "env"
	RevisionKey     = "revision"

	ControllerNameKey     = "controller"
	ResourceIdentifierKey = "obj"
	WorkloadIdentifierKey = "identity"
	NamespaceKey          = "namespace"
	LabelsKey             = "labels"
	LabelSelectorKey      = "labelSelector"
	ClusterKey            = "cluster"
	NamespaceSecretKey    = "namespace/secret"
	FilterNameKey         = "filterName"
	FilterTypeKey         = "filterType"
	HandlerNameKey        = "handlerName"

	DependentIdentityKey = "dependentIdentity"
	DependentsKey        = "dependents"
	SourceAssetKey       = "sourceAsset"
	DestinationAssetKey  = "destinationAsset"

	DeployNameKey  = "deployName"
	RolloutNameKey = "rolloutName"

	RouteNameKey = "routeName"
)
