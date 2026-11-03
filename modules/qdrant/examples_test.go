package qdrant_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func ExampleRun() {
	// runQdrantContainer {
	ctx := context.Background()

	qdrantContainer, err := qdrant.Run(ctx, "qdrant/qdrant:v1.7.4")
	defer func() {
		if err := testcontainers.TerminateContainer(qdrantContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := qdrantContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
