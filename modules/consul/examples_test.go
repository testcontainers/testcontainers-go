package consul_test

import (
	"context"
	"fmt"
	"log"

	capi "github.com/hashicorp/consul/api"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/consul"
)

func ExampleRun() {
	// runConsulContainer {
	ctx := context.Background()

	consulContainer, err := consul.Run(ctx, "hashicorp/consul:1.15")
	defer func() {
		if err := testcontainers.TerminateContainer(consulContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := consulContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	// connectConsul {
	ctx := context.Background()

	consulContainer, err := consul.Run(ctx, "hashicorp/consul:1.15")
	defer func() {
		if err := testcontainers.TerminateContainer(consulContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	endpoint, err := consulContainer.ApiEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get endpoint: %s", err)
		return
	}

	config := capi.DefaultConfig()
	config.Address = endpoint
	client, err := capi.NewClient(config)
	if err != nil {
		log.Printf("failed to connect to Consul: %s", err)
		return
	}
	// }

	nodeName, err := client.Agent().NodeName()
	if err != nil {
		log.Printf("failed to get node name: %s", err)
		return
	}
	fmt.Println(len(nodeName) > 0)

	// Output:
	// true
}
