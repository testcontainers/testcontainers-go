package pulsar_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/pulsar"
)

func ExampleRun() {
	// runPulsarContainer {
	ctx := context.Background()

	pulsarContainer, err := pulsar.Run(ctx, "docker.io/apachepulsar/pulsar:2.10.2")
	defer func() {
		if err := testcontainers.TerminateContainer(pulsarContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := pulsarContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
