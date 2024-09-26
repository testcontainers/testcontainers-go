package network_test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
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

func TestNew(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx,
		network.WithAttachable(),
		network.WithDriver("bridge"),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, net.Remove(ctx))
	}()

	networkName := net.Name

	nginxC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
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
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), dockernetwork.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	newNetwork := resources[0]
	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["this-is-a-test"] = "true"

	require.True(t, newNetwork.Attachable)
	require.False(t, newNetwork.Internal)
	require.Equal(t, "value", newNetwork.Labels["this-is-a-test"])
	require.NoError(t, testcontainers.TerminateContainer(nginxC))
}

// testNetworkAliases {
func TestContainerAttachedToNewNetwork(t *testing.T) {
	ctx := context.Background()

	newNetwork, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, newNetwork)

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

	newNetwork, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, newNetwork)

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
		n, err := network.New(ctx)
		require.NoError(t, err)
		testcontainers.CleanupNetwork(t, n)

		networks = append(networks, n.Name)
	}

	nginx, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        nginxAlpineImage,
			ExposedPorts: []string{nginxDefaultPort},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort(nginxDefaultPort),
				wait.ForLog("Configuration complete; ready for start up"),
			),
			Networks: networks,
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, nginx)
	require.NoError(t, err)

	containerId := nginx.GetContainerID()

	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)
	defer cli.Close()

	cnt, err := cli.ContainerInspect(ctx, containerId)
	require.NoError(t, err)
	require.Len(t, cnt.NetworkSettings.Networks, maxNetworksCount)
	require.NotNil(t, cnt.NetworkSettings.Networks[networks[0]])
	require.NotNil(t, cnt.NetworkSettings.Networks[networks[1]])
}

func TestMultipleContainersInTheNewNetwork(t *testing.T) {
	ctx := context.Background()

	net, err := network.New(ctx, network.WithDriver("bridge"))
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

	networkName := net.Name

	c1, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:    nginxAlpineImage,
			Networks: []string{networkName},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, c1)
	require.NoError(t, err)

	c2, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:    nginxAlpineImage,
			Networks: []string{networkName},
		},
		Started: true,
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
		network.WithIPAM(&ipamConfig),
		network.WithAttachable(),
		network.WithDriver("bridge"),
	)
	// }
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

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
	testcontainers.CleanupContainer(t, nginx)
	require.NoError(t, err)

	provider, err := testcontainers.ProviderDocker.GetProvider()
	require.NoError(t, err)
	defer provider.Close()

	//nolint:staticcheck
	foundNetwork, err := provider.GetNetwork(ctx, testcontainers.NetworkRequest{Name: networkName})
	require.NoError(t, err)
	require.Equal(t, ipamConfig, foundNetwork.IPAM)
}

func TestWithNetwork(t *testing.T) {
	// first create the network to be reused
	nw, err := network.New(context.Background(), network.WithLabels(map[string]string{"network-type": "unique"}))
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, nw)

	networkName := nw.Name

	// check that the network is reused, multiple times
	for i := 0; i < 100; i++ {
		req := testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{},
		}

		err := network.WithNetwork([]string{"alias"}, nw)(&req)
		require.NoError(t, err)

		require.Len(t, req.Networks, 1)
		require.Equal(t, networkName, req.Networks[0])

		require.Len(t, req.NetworkAliases, 1)
		require.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)
	}

	// verify that the network is created only once
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), dockernetwork.ListOptions{
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

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: nginxAlpineImage,
		},
	}

	err := network.WithNetwork([]string{"alias"}, nw)(&req)
	require.NoError(t, err)

	require.Len(t, req.Networks, 1)
	require.Equal(t, networkName, req.Networks[0])

	require.Len(t, req.NetworkAliases, 1)
	require.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	// verify that the network is created only once
	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), dockernetwork.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)
	require.Empty(t, resources) // no Docker network was created

	c, err := testcontainers.GenericContainer(context.Background(), req)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestWithNewNetwork(t *testing.T) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	err := network.WithNewNetwork(context.Background(), []string{"alias"},
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.NoError(t, err)
	require.Len(t, req.Networks, 1)

	networkName := req.Networks[0]

	require.Len(t, req.NetworkAliases, 1)
	require.Equal(t, map[string][]string{networkName: {"alias"}}, req.NetworkAliases)

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), dockernetwork.ListOptions{
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
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := network.WithNewNetwork(ctx, []string{"alias"},
		network.WithAttachable(),
		network.WithInternal(),
		network.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)(&req)
	require.Error(t, err)

	// we do not want to fail, just skip the network creation
	require.Empty(t, req.Networks)
	require.Empty(t, req.NetworkAliases)
}
