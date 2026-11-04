package artemis_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func ExampleRun() {
	// runArtemisContainer {
	ctx := context.Background()

	artemisContainer, err := artemis.Run(ctx,
		"apache/activemq-artemis:2.30.0",
		artemis.WithCredentials("test", "test"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(artemisContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := artemisContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
