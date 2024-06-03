package testcontainers_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/network"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testNetworkAliases {
func TestNewAttachedToNewNetwork(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := testcontainers.NewNetwork(ctx, tcnetwork.WithCheckDuplicate())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

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

	nginx, err := testcontainers.New(ctx, req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, nginx.Terminate(ctx))
	}()

	networks, err := nginx.Networks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networks) != 1 {
		t.Errorf("Expected networks 1. Got '%d'.", len(networks))
	}
	network := networks[0]
	if network != networkName {
		t.Errorf("Expected network name '%s'. Got '%s'.", networkName, network)
	}

	networkAliases, err := nginx.NetworkAliases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkAliases) != 1 {
		t.Errorf("Expected network aliases for 1 network. Got '%d'.", len(networkAliases))
	}

	networkAlias := networkAliases[networkName]

	require.NotEmpty(t, networkAlias)

	for _, alias := range aliases {
		require.Contains(t, networkAlias, alias)
	}

	networkIP, err := nginx.ContainerIP(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(networkIP) == 0 {
		t.Errorf("Expected an IP address, got %v", networkIP)
	}
}

// }

func TestContainerIPs(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := testcontainers.NewNetwork(ctx, tcnetwork.WithCheckDuplicate())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	networkName := newNetwork.Name

	nginx, err := testcontainers.New(ctx, testcontainers.Request{
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
	require.NoError(t, err)
	defer func() {
		require.NoError(t, nginx.Terminate(ctx))
	}()

	ips, err := nginx.ContainerIPs(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(ips) != 2 {
		t.Errorf("Expected two IP addresses, got %v", len(ips))
	}
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
		// use t.Cleanup to run after container.TerminateContainerOnEnd
		t.Cleanup(func() {
			require.NoError(t, n.Remove(ctx))
		})

		networks = append(networks, n.Name)
	}

	req := testcontainers.Request{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nginxDefaultPort),
			wait.ForLog("Configuration complete; ready for start up"),
		),
		Networks: networks,
		Started:  true,
	}

	nginx, err := testcontainers.New(ctx, req)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, nginx.Terminate(ctx))
	}()

	jsonRes, err := nginx.Inspect(ctx)
	require.NoError(t, err)

	assert.Len(t, jsonRes.NetworkSettings.Networks, maxNetworksCount)
	assert.NotNil(t, jsonRes.NetworkSettings.Networks[networks[0]])
	assert.NotNil(t, jsonRes.NetworkSettings.Networks[networks[1]])
}

func TestMultipleContainersInTheNewNetwork(t *testing.T) {
	ctx := context.Background()

	net, err := testcontainers.NewNetwork(ctx, tcnetwork.WithCheckDuplicate(), tcnetwork.WithDriver("bridge"))
	if err != nil {
		t.Fatal("cannot create network")
	}
	defer func() {
		require.NoError(t, net.Remove(ctx))
	}()

	networkName := net.Name

	req1 := testcontainers.Request{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
		Started:  true,
	}

	c1, err := testcontainers.New(ctx, req1)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		require.NoError(t, c1.Terminate(ctx))
	}()

	req2 := testcontainers.Request{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
		Started:  true,
	}
	c2, err := testcontainers.New(ctx, req2)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer func() {
		require.NoError(t, c2.Terminate(ctx))
	}()

	pNets, err := c1.Networks(ctx)
	require.NoError(t, err)

	rNets, err := c2.Networks(ctx)
	require.NoError(t, err)

	assert.Len(t, pNets, 1)
	assert.Len(t, rNets, 1)

	assert.Equal(t, networkName, pNets[0])
	assert.Equal(t, networkName, rNets[0])
}

func TestWithNetwork(t *testing.T) {
	// first create the network to be reused
	nw, err := testcontainers.NewNetwork(context.Background(), tcnetwork.WithCheckDuplicate(), tcnetwork.WithLabels(map[string]string{"network-type": "unique"}))
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

	args := filters.NewArgs()
	args.Add("name", networkName)

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Len(t, resources, 1)

	newNetwork := resources[0]

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["network-type"] = "unique"

	assert.Equal(t, networkName, newNetwork.Name)
	assert.False(t, newNetwork.Attachable)
	assert.False(t, newNetwork.Internal)
	assert.Equal(t, expectedLabels, newNetwork.Labels)
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

	assert.Len(t, req.Networks, 1)
	assert.Equal(t, networkName, req.Networks[0])

	assert.Len(t, req.NetworkAliases, 1)
	assert.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	// verify that the network is NOT created at all
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	args := filters.NewArgs()
	args.Add("name", networkName)

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Empty(t, resources) // no Docker network was created

	c, err := testcontainers.New(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, c)
	defer func() {
		require.NoError(t, c.Terminate(context.Background()))
	}()
}

func TestWithNewNetwork(t *testing.T) {
	req := testcontainers.Request{}

	err := testcontainers.WithNewNetwork(context.Background(), []string{"alias"},
		tcnetwork.WithAttachable(),
		tcnetwork.WithInternal(),
		tcnetwork.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.NoError(t, err)

	assert.Len(t, req.Networks, 1)

	networkName := req.Networks[0]

	assert.Len(t, req.NetworkAliases, 1)
	assert.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	args := filters.NewArgs()
	args.Add("name", networkName)

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Len(t, resources, 1)

	newNetwork := resources[0]
	defer func() {
		require.NoError(t, client.NetworkRemove(context.Background(), newNetwork.ID))
	}()

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["this-is-a-test"] = "value"

	assert.Equal(t, networkName, newNetwork.Name)
	assert.True(t, newNetwork.Attachable)
	assert.True(t, newNetwork.Internal)
	assert.Equal(t, expectedLabels, newNetwork.Labels)
}

func TestWithNewNetworkContextTimeout(t *testing.T) {
	req := testcontainers.Request{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := testcontainers.WithNewNetwork(ctx, []string{"alias"},
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.Error(t, err)

	// we do not want to fail, just skip the network creation
	assert.Empty(t, req.Networks)
	assert.Empty(t, req.NetworkAliases)
}
