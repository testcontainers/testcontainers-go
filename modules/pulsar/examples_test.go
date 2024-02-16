package pulsar_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/pulsar"
)

func ExampleRunContainer() {
	// runPulsarContainer {
	ctx := context.Background()

	pulsarContainer, err := pulsar.RunContainer(ctx,
		testcontainers.WithImage("docker.io/apachepulsar/pulsar:2.10.2"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := pulsarContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := pulsarContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
