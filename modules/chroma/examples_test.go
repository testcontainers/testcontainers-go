package chroma_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func ExampleRunContainer() {
	// runChromaContainer {
	ctx := context.Background()

	chromaContainer, err := chroma.RunContainer(ctx, testcontainers.WithImage("chromadb/chroma:0.4.22.dev44"))
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := chromaContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := chromaContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
