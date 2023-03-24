package testcontainers

import (
	"os"
	"testing"
)

// TestProviderType_GetProvider_autodetect should NOT be run in parallel as it sets environment variables.
func TestProviderType_GetProvider_autodetect(t *testing.T) {
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
			DockerHost: "unix:///var/run/docker.sock",
			want:       Bridge,
		},
		{
			name:       "default provider with podman.socket",
			tr:         ProviderDefault,
			DockerHost: "unix://$XDG_RUNTIME_DIR/podman/podman.sock",
			want:       Podman,
		},
		{
			name:       "docker provider without podman.socket",
			tr:         ProviderDocker,
			DockerHost: "unix:///var/run/docker.sock",
			want:       Bridge,
		},
		{
			// Explicitly setting Docker provider should not be overridden by auto-detect
			name:       "docker provider with podman.socket",
			tr:         ProviderDocker,
			DockerHost: "unix://$XDG_RUNTIME_DIR/podman/podman.sock",
			want:       Bridge,
		},
		{
			name:       "Podman provider without podman.socket",
			tr:         ProviderPodman,
			DockerHost: "unix:///var/run/docker.sock",
			want:       Podman,
		},
		{
			name:       "Podman provider with podman.socket",
			tr:         ProviderPodman,
			DockerHost: "unix://$XDG_RUNTIME_DIR/podman/podman.sock",
			want:       Podman,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := os.Getenv("DOCKER_HOST")
			defer os.Setenv("DOCKER_HOST", env)

			if err := os.Setenv("DOCKER_HOST", tt.DockerHost); err != nil {
				t.Fatalf("os.Setenv() = %v", err)
			}

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
