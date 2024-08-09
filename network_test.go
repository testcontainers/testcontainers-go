package testcontainers_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

func TestNew(t *testing.T) {
	ctx := context.Background()

	net, err := testcontainers.NewNetwork(ctx,
		tcnetwork.WithAttachable(),
		// Makes the network internal only, meaning the host machine cannot access it.
		// Remove or use `network.WithDriver("bridge")` to change the network's mode.
		tcnetwork.WithInternal(),
		tcnetwork.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)
	require.NoError(t, err)
	defer func() {
		if err := net.Remove(ctx); err != nil {
			t.Fatalf("failed to remove network: %s", err)
		}
	}()

	networkName := net.Name

	nginxC, _ := testcontainers.Run(ctx, testcontainers.Request{
		Image: "nginx:alpine",
		ExposedPorts: []string{
			"80/tcp",
		},
		Networks: []string{
			networkName,
		},
		Started: true,
	})
	defer func() {
		if err := nginxC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	client, err := testcontainers.NewDockerClientWithOpts(context.Background())
	require.NoError(t, err)

	resources, err := client.NetworkList(context.Background(), dockernetwork.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	require.NoError(t, err)

	assert.Len(t, resources, 1)

	newNetwork := resources[0]

	expectedLabels := testcontainers.GenericLabels()
	expectedLabels["this-is-a-test"] = "true"

	assert.True(t, newNetwork.Attachable)
	assert.True(t, newNetwork.Internal)
	assert.Equal(t, "value", newNetwork.Labels["this-is-a-test"])

	require.NoError(t, err)
}

func TestNewNetwork_withOptions(t *testing.T) {
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
	net, err := testcontainers.NewNetwork(ctx,
		tcnetwork.WithIPAM(&ipamConfig),
		tcnetwork.WithAttachable(),
		tcnetwork.WithDriver("bridge"),
	)
	// }
	if err != nil {
		t.Fatal("cannot create network: ", err)
	}
	defer func() {
		require.NoError(t, net.Remove(ctx))
	}()

	networkName := net.Name

	foundNetwork, err := corenetwork.GetByName(ctx, networkName)
	if err != nil {
		t.Fatal("Cannot get created network by name")
	}
	assert.Equal(t, ipamConfig, foundNetwork.IPAM)
}
