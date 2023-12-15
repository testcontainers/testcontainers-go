package network_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
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

	net, err := network.New(ctx, network.WithCheckDuplicate())
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

	nginxC.GetContainerID()

	state, err := nginxC.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
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
	if testcontainersdocker.IsWindows() {
		t.Skip("Skip for Windows. See https://stackoverflow.com/questions/43784916/docker-for-windows-networking-container-with-multiple-network-interfaces")
	}

	ctx := context.Background()
	networks := []string{}

	maxNetworksCount := 2

	for i := 0; i < maxNetworksCount; i++ {
		n, err := network.New(ctx)
		assert.Nil(t, err)
		// use t.Cleanup to run after terminateContainerOnEnd
		t.Cleanup(func() {
			assert.NoError(t, n.Remove(ctx))
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
	assert.Nil(t, err)
	defer cli.Close()

	cnt, err := cli.ContainerInspect(ctx, containerId)
	assert.Nil(t, err)
	assert.Equal(t, maxNetworksCount, len(cnt.NetworkSettings.Networks))
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

	assert.Equal(t, 1, len(pNets))
	assert.Equal(t, 1, len(rNets))

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
