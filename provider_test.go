package testcontainers

import (
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

func TestProviderTypeGetProviderAutodetect(t *testing.T) {
	var dockerSocket = testcontainersdocker.DockerSocketPathWithSchema
	const podmanSocket = "unix://$XDG_RUNTIME_DIR/podman/podman.sock"

	tests := []struct {
		name       string
		tr         ProviderType
		DockerHost string
		want       string
		wantErr    bool
	}{
		{
			name:       "docker provider with docker.socket",
			tr:         ProviderDocker,
			DockerHost: dockerSocket,
			want:       Bridge,
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
			if tt.tr == ProviderPodman && !testcontainersdocker.IsPodman() {
				t.Skip("Skipping podman not available")
			}
			if tt.tr == ProviderDocker && !testcontainersdocker.IsDocker() {
				t.Skip("Skipping because docker not available")
			}

			t.Setenv("DOCKER_HOST", tt.DockerHost)

			got, err := tt.tr.GetProvider()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProviderType.GetProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			provider, ok := got.(*DockerProvider)
			if !ok {
				t.Fatalf("ProviderType.GetProvider() = %T, want %T", got, &DockerProvider{})
			}

			defaultBridgeNetworkName := provider.BridgeNetworkName()
			if defaultBridgeNetworkName != tt.want {
				t.Errorf("ProviderType.GetProvider() = %v, want %v", defaultBridgeNetworkName, tt.want)
			}
		})
	}
}
