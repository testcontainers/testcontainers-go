package testcontainers

import (
	"context"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

func TestProviderTypeGetProviderAutodetect(t *testing.T) {
	dockerHost := testcontainersdocker.ExtractDockerHost(context.Background())
	const podmanSocket = "unix://$XDG_RUNTIME_DIR/podman/podman.sock"

	tests := []struct {
		name               string
		tr                 ProviderType
		PropertiesProvider string
		DockerHost         string
		want               string
		wantErr            bool
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
		{
			name:               "default provider with podman configured in properties",
			tr:                 ProviderDefault,
			PropertiesProvider: "podman",
			DockerHost:         dockerHost,
			want:               Podman,
		},
		{
			// Explicitly setting Docker provider should not be overridden by properties
			name:               "docker provider with podman configured in properties",
			tr:                 ProviderDocker,
			PropertiesProvider: "podman",
			DockerHost:         dockerHost,
			want:               Bridge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tr == ProviderPodman && testcontainersdocker.IsWindows() {
				t.Skip("Podman provider is not implemented for Windows")
			}

			t.Setenv("DOCKER_HOST", tt.DockerHost)

			setupTestcontainersProperties(t, "provider.type="+tt.PropertiesProvider)

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

func setupTestcontainersProperties(t *testing.T, content string) {
	t.Cleanup(func() {
		// reset the properties file after the test
		config.Reset()
	})

	config.Reset()

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	err := createTmpDir(homeDir)
	if err != nil {
		t.Fatalf("failed to create tmp home dir: %v", err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // Windows support

	if err := os.WriteFile(filepath.Join(homeDir, ".testcontainers.properties"), []byte(content), 0o600); err != nil {
		t.Errorf("Failed to create the file: %v", err)
		return
	}
}

func createTmpDir(dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	return nil
}
