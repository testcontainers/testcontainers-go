package testcontainers_test

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestProviderTypeGetProviderAutodetect(t *testing.T) {
	dockerHost := core.ExtractDockerHost(context.Background())
	const podmanSocket = "unix://$XDG_RUNTIME_DIR/podman/podman.sock"

	tests := []struct {
		name       string
		tr         testcontainers.ProviderType
		DockerHost string
		want       string
		wantErr    bool
	}{
		{
			name:       "default provider without podman.socket",
			tr:         testcontainers.ProviderDefault,
			DockerHost: dockerHost,
			want:       testcontainers.Bridge,
		},
		{
			name:       "default provider with podman.socket",
			tr:         testcontainers.ProviderDefault,
			DockerHost: podmanSocket,
			want:       testcontainers.Podman,
		},
		{
			name:       "docker provider without podman.socket",
			tr:         testcontainers.ProviderDocker,
			DockerHost: dockerHost,
			want:       testcontainers.Bridge,
		},
		{
			// Explicitly setting Docker provider should not be overridden by auto-detect
			name:       "docker provider with podman.socket",
			tr:         testcontainers.ProviderDocker,
			DockerHost: podmanSocket,
			want:       testcontainers.Bridge,
		},
		{
			name:       "Podman provider without podman.socket",
			tr:         testcontainers.ProviderPodman,
			DockerHost: dockerHost,
			want:       testcontainers.Podman,
		},
		{
			name:       "Podman provider with podman.socket",
			tr:         testcontainers.ProviderPodman,
			DockerHost: podmanSocket,
			want:       testcontainers.Podman,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tr == testcontainers.ProviderPodman && core.IsWindows() {
				t.Skip("Podman provider is not implemented for Windows")
			}

			t.Setenv("DOCKER_HOST", tt.DockerHost)

			got, err := tt.tr.GetProvider()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProviderType.GetProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			provider, ok := got.(*testcontainers.DockerProvider)
			if !ok {
				t.Fatalf("ProviderType.GetProvider() = %T, want %T", got, &testcontainers.DockerProvider{})
			}
			if provider.DefaultBridgeNetworkName != tt.want {
				t.Errorf("ProviderType.GetProvider() = %v, want %v", provider.DefaultBridgeNetworkName, tt.want)
			}
		})
	}
}
