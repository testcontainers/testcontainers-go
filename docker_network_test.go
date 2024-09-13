package testcontainers_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testNetworkAliases {
func TestNewAttachedToNewNetwork(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := testcontainers.NewNetwork(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, newNetwork)

	networkName := newNetwork.Name

	aliases := []string{"alias1", "alias2", "alias3"}

	req := testcontainers.Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: aliases,
		},
		Started: true,
	}

	nginx, err := testcontainers.Run(ctx, req)
	testcontainers.CleanupContainer(t, nginx)
	require.NoError(t, err)

	networks, err := nginx.Networks(ctx)
	require.NoError(t, err)
	require.Len(t, networks, 1)

	nw := networks[0]
	require.Equal(t, networkName, nw)

	networkAliases, err := nginx.NetworkAliases(ctx)
	require.NoError(t, err)

	require.Len(t, networkAliases, 1)

	networkAlias := networkAliases[networkName]
	require.NotEmpty(t, networkAlias)

	for _, alias := range aliases {
		require.Contains(t, networkAlias, alias)
	}

	networkIP, err := nginx.ContainerIP(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, networkIP)
}

// }

func TestContainerIPs(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := testcontainers.NewNetwork(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, newNetwork)

	networkName := newNetwork.Name

	nginx, err := testcontainers.Run(ctx, testcontainers.Request{
		Image: nginxAlpineImage,
		ExposedPorts: []string{
			nginxDefaultPort,
		},
		Networks: []string{
			"bridge",
			networkName,
		},
		WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		Started:    true,
	})
	testcontainers.CleanupContainer(t, nginx)
	require.NoError(t, err)

	ips, err := nginx.ContainerIPs(ctx)
	require.NoError(t, err)
	require.Len(t, ips, 2)
}

func TestContainerWithReaperNetwork(t *testing.T) {
	if core.IsWindows() {
		t.Skip("Skip for Windows. See https://stackoverflow.com/questions/43784916/docker-for-windows-networking-container-with-multiple-network-interfaces")
	}

	ctx := context.Background()
	networks := []string{}

	maxNetworksCount := 2

	for i := 0; i < maxNetworksCount; i++ {
		n, err := testcontainers.NewNetwork(ctx)
		require.NoError(t, err)
		testcontainers.CleanupNetwork(t, n)

		networks = append(networks, n.Name)
	}

	nginx, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nginxDefaultPort),
			wait.ForLog("Configuration complete; ready for start up"),
		),
		Networks: networks,
		Started:  true,
	})
	testcontainers.CleanupContainer(t, nginx)
	require.NoError(t, err)

	jsonRes, err := nginx.Inspect(ctx)
	require.NoError(t, err)
	require.Len(t, jsonRes.NetworkSettings.Networks, maxNetworksCount)
	require.NotNil(t, jsonRes.NetworkSettings.Networks[networks[0]])
	require.NotNil(t, jsonRes.NetworkSettings.Networks[networks[1]])
}

func TestMultipleContainersInTheNewNetwork(t *testing.T) {
	ctx := context.Background()

	net, err := testcontainers.NewNetwork(ctx, tcnetwork.WithDriver("bridge"))
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

	networkName := net.Name

	c1, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
		Started:  true,
	})
	testcontainers.CleanupContainer(t, c1)
	require.NoError(t, err)

	c2, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
		Started:  true,
	})
	testcontainers.CleanupContainer(t, c2)
	require.NoError(t, err)

	pNets, err := c1.Networks(ctx)
	require.NoError(t, err)

	rNets, err := c2.Networks(ctx)
	require.NoError(t, err)

	require.Len(t, pNets, 1)
	require.Len(t, rNets, 1)

	require.Equal(t, networkName, pNets[0])
	require.Equal(t, networkName, rNets[0])
}

func TestWithNetwork(t *testing.T) {
	// first create the network to be reused
	nw, err := testcontainers.NewNetwork(context.Background(), tcnetwork.WithLabels(map[string]string{"network-type": "unique"}))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, nw.Remove(context.Background()))
	}()

	networkName := nw.Name

	// check that the network is reused, multiple times
	for i := 0; i < 100; i++ {
		req := testcontainers.Request{}

		err := testcontainers.WithNetwork([]string{"alias"}, nw)(&req)
		require.NoError(t, err)

		assert.Len(t, req.Networks, 1)
		assert.Equal(t, networkName, req.Networks[0])

		assert.Len(t, req.NetworkAliases, 1)
		assert.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)
	}

	// verify that the network is created only once
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	newNetwork := resources[0]

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["network-type"] = "unique"

	require.Equal(t, networkName, newNetwork.Name)
	require.False(t, newNetwork.Attachable)
	require.False(t, newNetwork.Internal)
	require.Equal(t, expectedLabels, newNetwork.Labels)
}

func TestWithSyntheticNetwork(t *testing.T) {
	nw := &testcontainers.DockerNetwork{
		Name: "synthetic-network",
	}

	networkName := nw.Name

	req := testcontainers.Request{
		Image: nginxAlpineImage,
	}

	err := testcontainers.WithNetwork([]string{"alias"}, nw)(&req)
	require.NoError(t, err)

	require.Len(t, req.Networks, 1)
	require.Equal(t, networkName, req.Networks[0])

	require.Len(t, req.NetworkAliases, 1)
	require.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	// verify that the network is NOT created at all
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)
	assert.Empty(t, resources) // no Docker network was created

	c, err := testcontainers.Run(context.Background(), req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestWithNewNetwork(t *testing.T) {
	req := testcontainers.Request{}

	err := testcontainers.WithNewNetwork(context.Background(), []string{"alias"},
		tcnetwork.WithAttachable(),
		tcnetwork.WithInternal(),
		tcnetwork.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.NoError(t, err)

	require.Len(t, req.Networks, 1)

	networkName := req.Networks[0]

	require.Len(t, req.NetworkAliases, 1)
	require.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	newNetwork := resources[0]
	defer func() {
		require.NoError(t, client.NetworkRemove(context.Background(), newNetwork.ID))
	}()

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["this-is-a-test"] = "value"

	require.Equal(t, networkName, newNetwork.Name)
	require.True(t, newNetwork.Attachable)
	require.True(t, newNetwork.Internal)
	require.Equal(t, expectedLabels, newNetwork.Labels)
}

func TestWithNewNetworkContextTimeout(t *testing.T) {
	req := testcontainers.Request{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := testcontainers.WithNewNetwork(ctx, []string{"alias"},
		tcnetwork.WithAttachable(),
		tcnetwork.WithInternal(),
		tcnetwork.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.Error(t, err)

	// we do not want to fail, just skip the network creation
	require.Empty(t, req.Networks)
	require.Empty(t, req.NetworkAliases)
}
