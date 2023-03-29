package localstack

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func generateContainerRequest() *LocalStackContainerRequest {
	return &LocalStackContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env:          map[string]string{},
			ExposedPorts: []string{},
		},
	}
}

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
		defer dockerProvider.Close()

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

func TestIsLegacyMode(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"foo", true},
		{"latest", false},
		{"0.10.0", true},
		{"0.10.999", true},
		{"0.11", false},
		{"0.11.2", false},
		{"0.12", false},
		{"1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := isLegacyMode(fmt.Sprintf("localstack/localstack:%s", tt.version))
			assert.Equal(t, tt.want, got, "runInLegacyMode() = %v, want %v", got, tt.want)
		})
	}
}

func TestStart(t *testing.T) {
	ctx := context.Background()

	// withoutNetwork {
	container, err := StartContainer(
		ctx,
		OverrideContainerRequest(testcontainers.ContainerRequest{
			Image: fmt.Sprintf("localstack/localstack:%s", defaultVersion),
		}),
	)
	// }

	t.Run("multiple services should be exposed using the same port", func(t *testing.T) {
		require.Nil(t, err)
		assert.NotNil(t, container)

		rawPorts, err := container.Ports(ctx)
		require.Nil(t, err)

		ports := 0
		// only one port is exposed among all the ports in the container
		for _, v := range rawPorts {
			if len(v) > 0 {
				ports++
			}
		}

		assert.Equal(t, 1, ports) // a single port is exposed
	})
}

func TestStartWithoutOverride(t *testing.T) {
	// noopOverrideContainerRequest {
	ctx := context.Background()

	container, err := StartContainer(
		ctx,
		NoopOverrideContainerRequest,
	)
	require.Nil(t, err)
	assert.NotNil(t, container)
	// }
}

func TestStartWithNetwork(t *testing.T) {
	// withNetwork {
	ctx := context.Background()

	nw, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: "localstack-network",
		},
	})
	require.Nil(t, err)
	assert.NotNil(t, nw)

	container, err := StartContainer(
		ctx,
		OverrideContainerRequest(testcontainers.ContainerRequest{
			Image:          "localstack/localstack:0.13.0",
			Env:            map[string]string{"SERVICES": "s3,sqs"},
			Networks:       []string{"localstack-network"},
			NetworkAliases: map[string][]string{"localstack-network": {"localstack"}},
		}),
	)
	require.Nil(t, err)
	assert.NotNil(t, container)
	// }

	networks, err := container.Networks(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, len(networks))
	require.Equal(t, "localstack-network", networks[0])
}
