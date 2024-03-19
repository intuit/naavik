package resolver

type SecretConfigResolver struct{}

func NewSecretConfigResolver() ConfigResolver {
	return &SecretConfigResolver{}
}

// Default implementation of K8sConfigResolver(key, value)
// Returns the secret value as is without any processing.
func (s *SecretConfigResolver) GetKubeConfig(_ string, secretValue []byte) ([]byte, error) {
	return secretValue, nil
}
