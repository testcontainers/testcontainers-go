package testcontainersdocker

import (
	"context"
	"strings"
	"sync"
)

// ProviderType is an enum for the possible providers
type ProviderType int

// possible provider types
const (
	ProviderDocker ProviderType = iota
	ProviderPodman
)

var providerType ProviderType
var providerTypeOnce sync.Once

// Provider returns the container provider, extracting it from the environment
// See ExtractDockerHost for more details.
func Provider() ProviderType {
	providerTypeOnce.Do(func() {
		providerType = initProvider(context.Background())
	})

	return providerType
}

func initProvider(ctx context.Context) ProviderType {
	host := ExtractDockerHost(context.Background())

	if strings.HasSuffix(host, "podman.sock") {
		return ProviderPodman
	}

	return ProviderDocker
}

// IsDocker returns if the container provider is Docker
func IsDocker() bool {
	return Provider() == ProviderDocker
}

// IsPodman returns if the container provider is Podman
func IsPodman() bool {
	return Provider() == ProviderPodman
}
