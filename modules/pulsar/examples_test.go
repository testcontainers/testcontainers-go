package pulsar_test

import (
	"context"
	"fmt"

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
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := pulsarContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := pulsarContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
