package qdrant_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func ExampleRunContainer() {
	// runQdrantContainer {
	ctx := context.Background()

	qdrantContainer, err := qdrant.RunContainer(ctx, testcontainers.WithImage("qdrant/qdrant:v1.7.4"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := qdrantContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := qdrantContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
