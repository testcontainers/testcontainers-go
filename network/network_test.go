package network_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	nginxAlpineImage = "docker.io/nginx:alpine"
	nginxDefaultPort = "80/tcp"
)

// Create a network.
func ExampleNew() {
	// createNetwork {
	ctx := context.Background()

	net, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			panic(err)
		}
	}()

	networkName := net.Name
	// }

	nginxC, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "nginx:alpine",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
		},
		Started: true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	args := filters.NewArgs()
	args.Add("name", networkName)

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(len(resources))

	newNetwork := resources[0]

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["this-is-a-test"] = "true"

	fmt.Println(newNetwork.Attachable)
	fmt.Println(newNetwork.Internal)
	fmt.Println(newNetwork.Labels["this-is-a-test"])

	state, err := nginxC.State(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// 1
	// true
	// true
	// value
	// true
}

// testNetworkAliases {
func TestContainerAttachedToNewNetwork(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	networkName := newNetwork.Name

	aliases := []string{"alias1", "alias2", "alias3"}

	gcr := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
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
		},
		Started: true,
	}

	nginx, err := testcontainers.GenericContainer(ctx, gcr)
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

	newNetwork, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		require.NoError(t, newNetwork.Remove(ctx))
	})

	networkName := newNetwork.Name

	nginx, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: nginxAlpineImage,
			ExposedPorts: []string{
				nginxDefaultPort,
			},
			Networks: []string{
				"bridge",
				networkName,
			},
			WaitingFor: wait.ForListeningPort(nginxDefaultPort),
		},
		Started: true,
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
		n, err := network.New(ctx)
		require.NoError(t, err)
		// use t.Cleanup to run after terminateContainerOnEnd
		t.Cleanup(func() {
			require.NoError(t, n.Remove(ctx))
		})

		networks = append(networks, n.Name)
	}

	req := testcontainers.ContainerRequest{
		Image:        nginxAlpineImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nginxDefaultPort),
			wait.ForLog("Configuration complete; ready for start up"),
		),
		Networks: networks,
	}

	nginx, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err)
	defer func() {
		require.NoError(t, nginx.Terminate(ctx))
	}()

	containerId := nginx.GetContainerID()

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer cli.Close()

	cnt, err := cli.ContainerInspect(ctx, containerId)
	require.NoError(t, err)
	assert.Len(t, cnt.NetworkSettings.Networks, maxNetworksCount)
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[0]])
	assert.NotNil(t, cnt.NetworkSettings.Networks[networks[1]])
}

func TestMultipleContainersInTheNewNetwork(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithCheckDuplicate(), network.WithDriver("bridge"))
	if err != nil {
		t.Fatal("cannot create network")
	}
	defer func() {
		require.NoError(t, net.Remove(ctx))
	}()

	networkName := net.Name

	req1 := testcontainers.ContainerRequest{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
	}

	c1, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req1,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		require.NoError(t, c1.Terminate(ctx))
	}()

	req2 := testcontainers.ContainerRequest{
		Image:    nginxAlpineImage,
		Networks: []string{networkName},
	}
	c2, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req2,
		Started:          true,
	})
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

func TestNew_withOptions(t *testing.T) {
	// newNetworkWithOptions {
	ctx := context.Background()

	// dockernetwork is the alias used for github.com/docker/docker/api/types/network
	ipamConfig := dockernetwork.IPAM{
		Driver: "default",
		Config: []dockernetwork.IPAMConfig{
			{
				Subnet:  "10.1.1.0/24",
				Gateway: "10.1.1.254",
			},
		},
		Options: map[string]string{
			"driver": "host-local",
		},
	}
	net, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithIPAM(&ipamConfig),
		network.WithAttachable(),
		network.WithDriver("bridge"),
	)
	// }
	if err != nil {
		t.Fatal("cannot create network: ", err)
	}
	defer func() {
		require.NoError(t, net.Remove(ctx))
	}()

	networkName := net.Name

	nginx, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "nginx:alpine",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
		},
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, nginx.Terminate(ctx))
	}()

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		t.Fatal("Cannot get Provider")
	}
	defer provider.Close()

	//nolint:staticcheck
	foundNetwork, err := provider.GetNetwork(ctx, testcontainers.NetworkRequest{Name: networkName})
	if err != nil {
		t.Fatal("Cannot get created network by name")
	}
	assert.Equal(t, ipamConfig, foundNetwork.IPAM)
}

func TestWithNetwork(t *testing.T) {
	// first create the network to be reused
	nw, err := network.New(context.Background(), network.WithCheckDuplicate(), network.WithLabels(map[string]string{"network-type": "unique"}))
	require.NoError(t, err)
	defer func() {
		require.NoError(t, nw.Remove(context.Background()))
	}()

	networkName := nw.Name

	// check that the network is reused, multiple times
	for i := 0; i < 100; i++ {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{},
		}

		network.WithNetwork([]string{"alias"}, nw)(&req)

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

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: nginxAlpineImage,
		},
	}

	network.WithNetwork([]string{"alias"}, nw)(&req)

	assert.Len(t, req.Networks, 1)
	assert.Equal(t, networkName, req.Networks[0])

	assert.Len(t, req.NetworkAliases, 1)
	assert.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	// verify that the network is created only once
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	args := filters.NewArgs()
	args.Add("name", networkName)

	resources, err := client.NetworkList(context.Background(), types.NetworkListOptions{
		Filters: args,
	})
	require.NoError(t, err)
	assert.Empty(t, resources) // no Docker network was created

	c, err := testcontainers.GenericContainer(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, c)
	defer func() {
		require.NoError(t, c.Terminate(context.Background()))
	}()
}

func TestWithNewNetwork(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	network.WithNewNetwork(context.Background(), []string{"alias"},
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)

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
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	network.WithNewNetwork(ctx, []string{"alias"},
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)

	// we do not want to fail, just skip the network creation
	assert.Empty(t, req.Networks)
	assert.Empty(t, req.NetworkAliases)
}
