package utils

import (
	"strings"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ResourceInterface interface {
	GetWorkloadIdentifier(objectMetadata metav1.ObjectMeta) string
	IsResourceMeshEnabled(objectMetadata metav1.ObjectMeta) bool
	IsResourceIgnored(objectMetadata metav1.ObjectMeta) bool
	IsSyncEnabled(objectMetadata metav1.ObjectMeta) bool
	GetEnv(objectMetadata metav1.ObjectMeta, name string, namespace string) string
}

var resource = resourceUtil{}

type resourceUtil struct{}

func ResourceUtil() ResourceInterface {
	return resource
}

func (resourceUtil) GetWorkloadIdentifier(objectMetadata metav1.ObjectMeta) string {
	identity := ""

	if objectMetadata.Labels != nil {
		identity = objectMetadata.Labels[options.GetWorkloadIdentityKey()]
	}
	if len(identity) == 0 && objectMetadata.Annotations != nil {
		identity = objectMetadata.Annotations[options.GetWorkloadIdentityKey()]
	}
	return identity
}

func (resourceUtil) GetEnv(objectMetadata metav1.ObjectMeta, name string, namespace string) string {
	env := ""

	if objectMetadata.Labels != nil {
		env = objectMetadata.Labels[options.GetEnvKey()]
	}
	if len(env) == 0 && objectMetadata.Annotations != nil {
		env = objectMetadata.Annotations[options.GetEnvKey()]
	}

	if len(env) == 0 {
		splitNamespace := strings.Split(namespace, "-")
		if len(splitNamespace) > 1 {
			env = splitNamespace[len(splitNamespace)-1]
		}
		logger.Log.Warnf("using deprecated approach to deduce env from namespace for deployment, name=%v in namespace=%v", name, namespace)
	}
	if len(env) == 0 {
		env = types.EnvDefault
	}
	return env
}

func (resourceUtil) IsResourceMeshEnabled(objectMetadata metav1.ObjectMeta) bool {
	if objectMetadata.Labels != nil {
		enabled := objectMetadata.Labels[options.GetMeshInjectionEnabledKey()]
		if strings.ToLower(enabled) == types.IsTrue {
			return true
		}
	}
	if objectMetadata.Annotations != nil {
		enabled := objectMetadata.Annotations[options.GetMeshInjectionEnabledKey()]
		if strings.ToLower(enabled) == types.IsTrue {
			return true
		}
	}
	return false
}

func (resourceUtil) IsResourceIgnored(objectMetadata metav1.ObjectMeta) bool {
	if objectMetadata.Labels != nil {
		ignored := objectMetadata.Labels[options.GetResourceIgnoreLabel()]
		if strings.ToLower(ignored) == types.IsTrue {
			return true
		}
	}
	if objectMetadata.Annotations != nil {
		ignored := objectMetadata.Annotations[options.GetResourceIgnoreLabel()]
		if strings.ToLower(ignored) == types.IsTrue {
			return true
		}
	}
	return false
}

func (resourceUtil) IsSyncEnabled(objectMetadata metav1.ObjectMeta) bool {
	if objectMetadata.Labels != nil {
		ignored := objectMetadata.Labels[options.GetSecretSyncLabel()]
		if strings.ToLower(ignored) == types.IsTrue {
			return true
		}
	}
	if objectMetadata.Annotations != nil {
		ignored := objectMetadata.Annotations[options.GetSecretSyncLabel()]
		if strings.ToLower(ignored) == types.IsTrue {
			return true
		}
	}
	return false
}
