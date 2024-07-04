package redpanda_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func ExampleRun() {
	// runRedpandaContainer {
	ctx := context.Background()

	redpandaContainer, err := redpanda.Run(ctx,
		"docker.redpanda.com/redpandadata/redpanda:v23.3.3",
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithEnableWasmTransform(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := redpandaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := redpandaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
