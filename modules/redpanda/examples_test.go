package redpanda_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func ExampleRunContainer() {
	// runRedpandaContainer {
	ctx := context.Background()

	redpandaContainer, err := redpanda.RunContainer(ctx,
		redpanda.WithEnableSASL(),
		redpanda.WithEnableKafkaAuthorization(),
		redpanda.WithNewServiceAccount("superuser-1", "test"),
		redpanda.WithNewServiceAccount("superuser-2", "test"),
		redpanda.WithNewServiceAccount("no-superuser", "test"),
		redpanda.WithSuperusers("superuser-1", "superuser-2"),
		redpanda.WithEnableSchemaRegistryHTTPBasicAuth(),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := redpandaContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := redpandaContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
