package consul_test

import (
	"context"
	"fmt"
	"log"

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
