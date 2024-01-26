package inbucket_test

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/inbucket"
)

func ExampleRunContainer() {
	// runInbucketContainer {
	ctx := context.Background()

	inbucketContainer, err := inbucket.RunContainer(ctx, testcontainers.WithImage("inbucket/inbucket:sha-2d409bb"))
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := inbucketContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := inbucketContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
