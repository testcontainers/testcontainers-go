package testcontainers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestProviderTypeGetUnderlyingProviderType(t *testing.T) {
	tests := []struct {
		name           string
		providerType   ProviderType
		propertiesFile string // content of .testcontainers.properties
		env            map[string]string
		expectedType   ProviderType
	}{
		{
			name:         "ProviderDocker always returns ProviderDocker",
			providerType: ProviderDocker,
			expectedType: ProviderDocker,
		},
		{
			name:         "ProviderPodman always returns ProviderPodman",
			providerType: ProviderPodman,
			expectedType: ProviderPodman,
		},
		{
			name:           "ProviderDefault with properties file set to docker",
			providerType:   ProviderDefault,
			propertiesFile: "provider=docker",
			expectedType:   ProviderDocker,
		},
		{
			name:           "ProviderDefault with properties file set to podman",
			providerType:   ProviderDefault,
			propertiesFile: "provider=podman",
			expectedType:   ProviderPodman,
		},
		{
			name:         "ProviderDefault with env var set to docker",
			providerType: ProviderDefault,
			env: map[string]string{
				"TESTCONTAINERS_PROVIDER": "docker",
			},
			expectedType: ProviderDocker,
		},
		{
			name:         "ProviderDefault with env var set to podman",
			providerType: ProviderDefault,
			env: map[string]string{
				"TESTCONTAINERS_PROVIDER": "podman",
			},
			expectedType: ProviderPodman,
		},
		{
			name:           "ProviderDefault with env var podman and properties docker - env wins",
			providerType:   ProviderDefault,
			propertiesFile: "provider=docker",
			env: map[string]string{
				"TESTCONTAINERS_PROVIDER": "podman",
			},
			expectedType: ProviderPodman,
		},
		{
			name:           "ProviderDocker with env var podman and properties podman - explicit provider wins",
			providerType:   ProviderDocker,
			propertiesFile: "provider=podman",
			env: map[string]string{
				"TESTCONTAINERS_PROVIDER": "podman",
			},
			expectedType: ProviderDocker,
		},
		{
			name:           "ProviderPodman with env var docker and properties docker - explicit provider wins",
			providerType:   ProviderPodman,
			propertiesFile: "provider=docker",
			env: map[string]string{
				"TESTCONTAINERS_PROVIDER": "docker",
			},
			expectedType: ProviderPodman,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset config for each test to ensure clean state
			config.Reset()

			// Create temp directory for HOME
			tmpDir := t.TempDir()
			t.Setenv("HOME", tmpDir)
			t.Setenv("USERPROFILE", tmpDir) // Windows support

			// Set any additional environment variables
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			// Create properties file if content is provided
			if tt.propertiesFile != "" {
				err := os.WriteFile(
					filepath.Join(tmpDir, ".testcontainers.properties"),
					[]byte(tt.propertiesFile),
					0600,
				)
				require.NoError(t, err, "Failed to create properties file")
			}

			// Test UnderlyingProviderType
			result := tt.providerType.UnderlyingProviderType()
			require.Equal(t, tt.expectedType, result, "UnderlyingProviderType() returned unexpected type")
		})
	}
}

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
