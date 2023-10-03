package k6_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k6"
)

func ExamplerunContainer() {
	// runk6Container {
	ctx := context.Background()

	k6Container, err := k6.runContainer(ctx, testcontainers.WithImage("grafana/k6"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := k6Container.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := k6Container.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
