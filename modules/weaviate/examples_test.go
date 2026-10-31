package weaviate_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	tcweaviate "github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func ExampleRun() {
	// runWeaviateContainer {
	ctx := context.Background()

	weaviateContainer, err := tcweaviate.Run(ctx, "semitechnologies/weaviate:1.29.0")
	defer func() {
		if err := testcontainers.TerminateContainer(weaviateContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := weaviateContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
