package testcontainers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestProviderTypeGetProviderAutodetect(t *testing.T) {
	dockerHost := core.MustExtractDockerHost(context.Background())
	const podmanSocket = "unix://$XDG_RUNTIME_DIR/podman/podman.sock"

	tests := []struct {
		name       string
		tr         ProviderType
		DockerHost string
		want       string
	}{
		{
			name:       "default provider without podman.socket",
			tr:         ProviderDefault,
			DockerHost: dockerHost,
			want:       Bridge,
		},
		{
			name:       "default provider with podman.socket",
			tr:         ProviderDefault,
			DockerHost: podmanSocket,
			want:       Podman,
		},
		{
			name:       "docker provider without podman.socket",
			tr:         ProviderDocker,
			DockerHost: dockerHost,
			want:       Bridge,
		},
		{
			// Explicitly setting Docker provider should not be overridden by auto-detect
			name:       "docker provider with podman.socket",
			tr:         ProviderDocker,
			DockerHost: podmanSocket,
			want:       Bridge,
		},
		{
			name:       "Podman provider without podman.socket",
			tr:         ProviderPodman,
			DockerHost: dockerHost,
			want:       Podman,
		},
		{
			name:       "Podman provider with podman.socket",
			tr:         ProviderPodman,
			DockerHost: podmanSocket,
			want:       Podman,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tr == ProviderPodman && core.IsWindows() {
				t.Skip("Podman provider is not implemented for Windows")
			}

			t.Setenv("DOCKER_HOST", tt.DockerHost)

			got, err := tt.tr.GetProvider()
			require.NoErrorf(t, err, "ProviderType.GetProvider()")
			provider, ok := got.(*DockerProvider)
			require.Truef(t, ok, "ProviderType.GetProvider() = %T, want %T", got, &DockerProvider{})
			require.Equalf(t, tt.want, provider.defaultBridgeNetworkName, "ProviderType.GetProvider() = %v, want %v", provider.defaultBridgeNetworkName, tt.want)
		})
	}
}
