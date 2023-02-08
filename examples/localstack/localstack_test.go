package localstack

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
)

func TestConfigureDockerHost(t *testing.T) {

	t.Run("HOSTNAME_EXTERNAL variable is passed as part of the request", func(t *testing.T) {
		req := generateContainerRequest()

		req.Env[hostnameExternalEnvVar] = "foo"

		reason, err := configureDockerHost(req)
		assert.Nil(t, err)
		assert.Equal(t, "explicitly as environment variable", reason)
	})

	t.Run("HOSTNAME_EXTERNAL matches the last network alias on a container with non-default network", func(t *testing.T) {
		req := generateContainerRequest()

		req.Networks = []string{"foo", "bar", "baaz"}
		req.NetworkAliases = map[string][]string{
			"foo":  {"foo0", "foo1", "foo2", "foo3"},
			"bar":  {"bar0", "bar1", "bar2", "bar3"},
			"baaz": {"baaz0", "baaz1", "baaz2", "baaz3"},
		}

		reason, err := configureDockerHost(req)
		assert.Nil(t, err)
		assert.Equal(t, "to match last network alias on container with non-default network", reason)
		assert.Equal(t, "foo3", req.Env[hostnameExternalEnvVar])
	})

	t.Run("HOSTNAME_EXTERNAL matches the daemon host because there are no aliases", func(t *testing.T) {
		dockerProvider, err := testcontainers.NewDockerProvider()
		assert.Nil(t, err)

		// because the daemon host could be a remote one, we need to get it from the provider
		expectedDaemonHost, err := dockerProvider.DaemonHost(context.Background())
		assert.Nil(t, err)

		req := generateContainerRequest()

		req.Networks = []string{"foo", "bar", "baaz"}
		req.NetworkAliases = map[string][]string{}

		reason, err := configureDockerHost(req)
		assert.Nil(t, err)
		assert.Equal(t, "to match host-routable address for container", reason)
		assert.Equal(t, expectedDaemonHost, req.Env[hostnameExternalEnvVar])
	})
}

func TestRunInLegacyMode(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"foo", true},
		{"latest", false},
		{"0.10.0", true},
		{"0.11", false},
		{"0.11.2", false},
		{"0.12", false},
		{"1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := runInLegacyMode(fmt.Sprintf("localstack/localstack:%s", tt.version))
			assert.Equal(t, tt.want, got, "runInLegacyMode() = %v, want %v", got, tt.want)
		})
	}
}
