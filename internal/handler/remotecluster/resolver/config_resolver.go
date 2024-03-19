package resolver

import (
	"github.com/intuit/naavik/internal/types"
	"github.com/intuit/naavik/internal/types/context"
)

type ConfigResolver interface {
	GetKubeConfig(secretKey string, secretValue []byte) ([]byte, error)
}

func GetConfigResolver(ctx context.Context, configResolver string) ConfigResolver {
	switch configResolver {
	case types.ConfigResolverSecret:
		return NewSecretConfigResolver()
	default:
		ctx.Log.Fatalf("invalid config resolver %q", configResolver)
	}
	return nil
}
