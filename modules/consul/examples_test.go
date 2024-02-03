package consul_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/consul"
)

func ExampleRunContainer() {
	// runconsulContainer {
	ctx := context.Background()

	consulContainer, err := consul.RunContainer(ctx,
		testcontainers.WithImage("docker.io/hashicorp/consul:1.15"),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := consulContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := consulContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
