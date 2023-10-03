package k6_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k6"
)

func ExampleRunContainer() {
	// runK6Container {
	ctx := context.Background()

	k6Container, err := k6.RunContainer(ctx, testcontainers.WithImage("szkiba/k6x"))
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
