package testcontainers

import (
	"testing"
)

func TestProviderTypeGetProviderAutodetect(t *testing.T) {
	const dockerSocket = "unix:///var/run/docker.sock"
	const podmanSocket = "unix://$XDG_RUNTIME_DIR/podman/podman.sock"

	tests := []struct {
		name       string
		tr         ProviderType
		DockerHost string
		want       string
		wantErr    bool
	}{
		{
			name:       "default provider without podman.socket",
			tr:         ProviderDefault,
			DockerHost: dockerSocket,
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
			DockerHost: dockerSocket,
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
			DockerHost: dockerSocket,
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
			if provider.defaultBridgeNetworkName != tt.want {
				t.Errorf("ProviderType.GetProvider() = %v, want %v", provider.defaultBridgeNetworkName, tt.want)
			}
		})
	}
}
