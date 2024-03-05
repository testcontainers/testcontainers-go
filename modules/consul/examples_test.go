package consul_test

import (
	"context"
	"fmt"
	"log"

	capi "github.com/hashicorp/consul/api"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/consul"
)

func ExampleRunContainer() {
	// runConsulContainer {
	ctx := context.Background()

	consulContainer, err := consul.RunContainer(ctx,
		testcontainers.WithImage("docker.io/hashicorp/consul:1.15"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := consulContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := consulContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	// connectConsul {
	ctx := context.Background()

	consulContainer, err := consul.RunContainer(ctx,
		testcontainers.WithImage("docker.io/hashicorp/consul:1.15"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := consulContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	endpoint, err := consulContainer.ApiEndpoint(ctx)
	if err != nil {
		log.Fatalf("failed to get endpoint: %s", err) // nolint:gocritic
	}

	config := capi.DefaultConfig()
	config.Address = endpoint
	client, err := capi.NewClient(config)
	if err != nil {
		log.Fatalf("failed to connect to Consul: %s", err)
	}
	// }

	node_name, err := client.Agent().NodeName()
	if err != nil {
		log.Fatalf("failed to get node name: %s", err) // nolint:gocritic
	}
	fmt.Println(len(node_name) > 0)

	// Output:
	// true
}
