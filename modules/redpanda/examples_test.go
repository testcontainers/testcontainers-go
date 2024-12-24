package redpanda_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
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
		redpanda.WithBootstrapConfig("data_transforms_per_core_memory_reservation", 33554432),
		redpanda.WithBootstrapConfig("data_transforms_per_function_memory_limit", 16777216),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(redpandaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := redpandaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
