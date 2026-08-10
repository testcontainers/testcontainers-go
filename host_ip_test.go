package testcontainers

import (
	"context"
	"net/netip"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreCreateContainerHookAppliesHostIP tests that the pre-create hook binds
// all exposed ports to the requested host IP, both for the default ephemeral
// bindings and for bindings set by a HostConfigModifier.
func TestPreCreateContainerHookAppliesHostIP(t *testing.T) {
	ctx := context.Background()
	p := &DockerProvider{}

	t.Run("applies to default ephemeral bindings", func(t *testing.T) {
		dockerInput := &container.Config{}
		hostConfig := &container.HostConfig{}
		networkingConfig := &network.NetworkingConfig{}

		req := ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp"},
			HostIP:       "127.0.0.1",
		}

		err := p.preCreateContainerHook(ctx, req, dockerInput, hostConfig, networkingConfig)
		require.NoError(t, err)

		port := network.MustParsePort("80/tcp")
		bindings := hostConfig.PortBindings[port]
		require.Len(t, bindings, 1)
		assert.Equal(t, "127.0.0.1", bindings[0].HostIP.String())
		assert.Equal(t, "0", bindings[0].HostPort, "HostPort should remain ephemeral")
	})

	t.Run("overrides bindings set by HostConfigModifier", func(t *testing.T) {
		dockerInput := &container.Config{}
		hostConfig := &container.HostConfig{}
		networkingConfig := &network.NetworkingConfig{}

		req := ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp", "443/tcp"},
			HostIP:       "127.0.0.1",
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.PortBindings = network.PortMap{
					network.MustParsePort("443/tcp"): {{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "8443"}},
				}
			},
		}

		err := p.preCreateContainerHook(ctx, req, dockerInput, hostConfig, networkingConfig)
		require.NoError(t, err)

		for _, port := range []network.Port{network.MustParsePort("80/tcp"), network.MustParsePort("443/tcp")} {
			bindings := hostConfig.PortBindings[port]
			require.Len(t, bindings, 1, "expected a binding for %s", port)
			assert.Equal(t, "127.0.0.1", bindings[0].HostIP.String(), "binding for %s should use the requested host IP", port)
		}

		bindings := hostConfig.PortBindings[network.MustParsePort("443/tcp")]
		assert.Equal(t, "8443", bindings[0].HostPort, "custom HostPort should be preserved")
	})

	t.Run("invalid IP returns error", func(t *testing.T) {
		dockerInput := &container.Config{}
		hostConfig := &container.HostConfig{}
		networkingConfig := &network.NetworkingConfig{}

		req := ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp"},
			HostIP:       "not-an-ip",
		}

		err := p.preCreateContainerHook(ctx, req, dockerInput, hostConfig, networkingConfig)
		require.Error(t, err)
	})

	t.Run("no-op when HostIP is empty", func(t *testing.T) {
		dockerInput := &container.Config{}
		hostConfig := &container.HostConfig{}
		networkingConfig := &network.NetworkingConfig{}

		req := ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp"},
		}

		err := p.preCreateContainerHook(ctx, req, dockerInput, hostConfig, networkingConfig)
		require.NoError(t, err)

		bindings := hostConfig.PortBindings[network.MustParsePort("80/tcp")]
		require.Len(t, bindings, 1)
		assert.Zero(t, bindings[0].HostIP, "HostIP should remain empty (bind all interfaces)")
	})
}

// TestContainerWithHostIP is an integration test verifying that a container
// started with a host IP binds its exposed ports to that address only.
func TestContainerWithHostIP(t *testing.T) {
	ctx := context.Background()

	t.Run("ContainerRequest.HostIP", func(t *testing.T) {
		req := ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			HostIP:       "127.0.0.1",
		}

		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)
		defer func() {
			require.NoError(t, container.Terminate(ctx))
		}()

		inspect, err := container.Inspect(ctx)
		require.NoError(t, err)

		port := network.MustParsePort(nginxDefaultPort)
		bindings, ok := inspect.NetworkSettings.Ports[port]
		require.True(t, ok, "expected a binding for %s", nginxDefaultPort)
		require.Len(t, bindings, 1)
		require.Equal(t, "127.0.0.1", bindings[0].HostIP.String())
	})

	t.Run("WithHostIP option", func(t *testing.T) {
		container, err := Run(ctx, nginxAlpineImage,
			WithExposedPorts(nginxDefaultPort),
			WithHostIP("127.0.0.1"),
		)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, container.Terminate(ctx))
		}()

		inspect, err := container.Inspect(ctx)
		require.NoError(t, err)

		port := network.MustParsePort(nginxDefaultPort)
		bindings, ok := inspect.NetworkSettings.Ports[port]
		require.True(t, ok, "expected a binding for %s", nginxDefaultPort)
		require.Len(t, bindings, 1)
		require.Equal(t, "127.0.0.1", bindings[0].HostIP.String())
	})
}
