package typesense_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/typesense"
)

func ExampleRun() {
	// runTypesenseContainer {
	ctx := context.Background()

	typesenseContainer, err := typesense.Run(
		ctx,
		"typesense/typesense:26.0",
		typesense.WithAPIKey("my-api-key"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(typesenseContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := typesenseContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)
	fmt.Println(typesenseContainer.APIKey())

	// Output:
	// true
	// my-api-key
}
