package testcontainers

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

// Create a network using a provider. By default it is Docker.
func ExampleNetworkProvider_CreateNetwork() {
	ctx := context.Background()
	networkName := "new-network"
	gcr := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image: "nginx",
			ExposedPorts: []string{
				"80/tcp",
			},
			Networks: []string{
				networkName,
			},
		},
	}
	provider, _ := gcr.ProviderType.GetProvider()
	net, _ := provider.CreateNetwork(ctx, NetworkRequest{
		Name:           networkName,
		CheckDuplicate: true,
	})
	defer net.Remove(ctx)

	nginxC, _ := GenericContainer(ctx, gcr)
	defer nginxC.Terminate(ctx)
	nginxC.GetContainerID()
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
		Image:        "postgres:12.2",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env:          env,
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Networks:     []string{networkName},
	}

	gcr := GenericContainerRequest{
		ContainerRequest: dbContainerRequest,
		Started:          true,
	}

	provider, err := gcr.ProviderType.GetProvider()
	if err != nil {
		t.Fatal("cannot get provider")
	}

	net, err := provider.CreateNetwork(ctx, networkRequest)
	if err != nil {
		t.Fatal("cannot create network")
	}

	defer net.Remove(ctx)

	postgres, err := GenericContainer(ctx, gcr)
	if err != nil {
		t.Fatal(err)
	}

	defer postgres.Terminate(ctx)

	env = make(map[string]string)
	env["RABBITMQ_ERLANG_COOKIE"] = "f2a2d3d27c75"
	env["RABBITMQ_DEFAULT_USER"] = "admin"
	env["RABBITMQ_DEFAULT_PASS"] = "Password1"
	hp := wait.ForListeningPort("5672/tcp")
	hp.WithStartupTimeout(3 * time.Minute)
	amqpRequest := ContainerRequest{
		Image:        "rabbitmq:management-alpine",
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

	defer rabbitmq.Terminate(ctx)
	fmt.Println(postgres.GetContainerID())
	fmt.Println(rabbitmq.GetContainerID())
}
