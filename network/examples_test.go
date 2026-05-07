package network_test

import (
	"context"
	"fmt"
	"log"
	"net/netip"

	dockernetwork "github.com/moby/moby/api/types/network"

	"github.com/testcontainers/testcontainers-go/network"
)

func ExampleNew() {
	// createNetwork {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()
	// }

	fmt.Println(net.ID != "")
	fmt.Println(net.Driver)

	// Output:
	// true
	// bridge
}

func ExampleNew_withOptions() {
	// newNetworkWithOptions {
	ctx := context.Background()

	// dockernetwork is the alias used for github.com/docker/docker/api/types/network
	ipamConfig := dockernetwork.IPAM{
		Driver: "default",
		Config: []dockernetwork.IPAMConfig{
			{
				Subnet:  netip.MustParsePrefix("10.1.1.0/24"),
				Gateway: netip.MustParseAddr("10.1.1.254"),
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
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := net.Remove(ctx); err != nil {
			log.Printf("failed to remove network: %s", err)
		}
	}()
	// }

	fmt.Println(net.ID != "")
	fmt.Println(net.Driver)

	// Output:
	// true
	// bridge
}
