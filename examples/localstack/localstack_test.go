package localstack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigure(t *testing.T) {

	t.Run("HOSTNAME_EXTERNAL variable is passed as part of the request", func(t *testing.T) {
		req := generateContainerRequest()

		req.Env[hostnameExternalEnvVar] = "foo"

		reason, err := configure(req)
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

		reason, err := configure(req)
		assert.Nil(t, err)
		assert.Equal(t, "to match last network alias on container with non-default network", reason)
		assert.Equal(t, "foo3", req.Env[hostnameExternalEnvVar])
	})

	t.Run("HOSTNAME_EXTERNAL matches localhost because there are no aliases", func(t *testing.T) {
		req := generateContainerRequest()

		req.Networks = []string{"foo", "bar", "baaz"}
		req.NetworkAliases = map[string][]string{}

		reason, err := configure(req)
		assert.Nil(t, err)
		assert.Equal(t, "to match host-routable address for container", reason)
		assert.Equal(t, "localhost", req.Env[hostnameExternalEnvVar])
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
			got := runInLegacyMode(tt.version)
			assert.Equal(t, tt.want, got, "runInLegacyMode() = %v, want %v", got, tt.want)
		})
	}
}

func TestLocalStack(t *testing.T) {
	ctx := context.Background()

	container, err := setupLocalStack(ctx, defaultVersion, false)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions
}
