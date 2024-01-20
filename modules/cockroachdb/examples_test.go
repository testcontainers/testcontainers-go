package cockroachdb_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func ExampleRunContainer() {
	// runCockroachDBContainer {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.RunContainer(ctx, testcontainers.WithImage("cockroachdb/cockroach:latest-v23.1"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := cockroachdbContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
