package testcontainers_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
)

// Create a network.
func ExampleNewNetwork() {
	// createNetwork {
	ctx := context.Background()

	net, err := testcontainers.NewNetwork(ctx,
		tcnetwork.WithCheckDuplicate(),
		tcnetwork.WithAttachable(),
		// Makes the network internal only, meaning the host machine cannot access it.
		// Remove or use `network.WithDriver("bridge")` to change the network's mode.
		tcnetwork.WithInternal(),
		tcnetwork.WithLabels(map[string]string{"this-is-a-test": "value"}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			log.Fatalf("failed to remove network: %s", err)
		}
	}()

	networkName := net.Name
	// }

	client, err := core.NewClient(context.Background())
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

	expectedLabels := core.DefaultLabels(core.SessionID())
	expectedLabels["this-is-a-test"] = "true"

	fmt.Println(newNetwork.Attachable)
	fmt.Println(newNetwork.Internal)
	fmt.Println(newNetwork.Labels["this-is-a-test"])

	// Output:
	// 1
	// true
	// true
	// value
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
		tcnetwork.WithCheckDuplicate(),
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
