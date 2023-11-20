package testcontainers

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/wait"
)

// Create a network using a provider. By default it is Docker.
func ExampleNetworkProvider_CreateNetwork() {
	// createNetwork {
	ctx := context.Background()
	networkName := "new-generic-network"
	net, _ := GenericNetwork(ctx, GenericNetworkRequest{
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	defer func() {
		if err := net.Remove(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
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

func Test_NetworkWithIPAM(t *testing.T) {
	// withIPAM {
	ctx := context.Background()
	networkName := "test-network-with-ipam"
	ipamConfig := network.IPAM{
		Driver: "default",
		Config: []network.IPAMConfig{
			{
				Subnet:  "10.1.1.0/24",
				Gateway: "10.1.1.254",
			},
		},
		Options: map[string]string{
			"driver": "host-local",
		},
	}
	net, err := GenericNetwork(ctx, GenericNetworkRequest{
		NetworkRequest: NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
			IPAM:           &ipamConfig,
		},
	})
	// }
	if err != nil {
		t.Fatal("cannot create network: ", err)
	}

	defer func() {
		_ = net.Remove(ctx)
	}()

	nginxC, _ := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
		},
	})
	terminateContainerOnEnd(t, ctx, nginxC)
	nginxC.GetContainerID()

	provider, err := ProviderDocker.GetProvider()
	if err != nil {
		t.Fatal("Cannot get Provider")
	}
	defer provider.Close()

	foundNetwork, err := provider.GetNetwork(ctx, NetworkRequest{Name: networkName})
	if err != nil {
		t.Fatal("Cannot get created network by name")
	}
	assert.Equal(t, ipamConfig, foundNetwork.IPAM)
}

func Test_MultipleContainersInTheNewNetwork(t *testing.T) {
	ctx := context.Background()

	networkName := "test-network"

	networkRequest := NetworkRequest{
		Driver:     "bridge",
		Name:       networkName,
		Attachable: true,
	}

	env := make(map[string]string)
	env["POSTGRES_PASSWORD"] = "Password1"
	dbContainerRequest := ContainerRequest{
		Image:        "postgres:12",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env:          env,
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Networks:     []string{networkName},
	}

	net, err := GenericNetwork(ctx, GenericNetworkRequest{
		NetworkRequest: networkRequest,
	})
	if err != nil {
		t.Fatal("cannot create network")
	}

	defer func() {
		_ = net.Remove(ctx)
	}()

	postgres, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: dbContainerRequest,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	terminateContainerOnEnd(t, ctx, postgres)

	env = make(map[string]string)
	env["RABBITMQ_ERLANG_COOKIE"] = "f2a2d3d27c75"
	env["RABBITMQ_DEFAULT_USER"] = "admin"
	env["RABBITMQ_DEFAULT_PASS"] = "Password1"
	hp := wait.ForListeningPort("5672/tcp")
	hp.WithStartupTimeout(3 * time.Minute)
	amqpRequest := ContainerRequest{
		Image:        "rabbitmq:3.8.19-management-alpine",
		ExposedPorts: []string{"15672/tcp", "5672/tcp"},
		Env:          env,
		AutoRemove:   true,
		WaitingFor:   hp,
		Networks:     []string{networkName},
	}
	rabbitmq, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: amqpRequest,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
		return
	}

	terminateContainerOnEnd(t, ctx, rabbitmq)
	fmt.Println(postgres.GetContainerID())
	fmt.Println(rabbitmq.GetContainerID())
}
