package opensearch_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go/modules/opensearch"
)

func ExampleRun() {
	// runOpenSearchContainer {
	ctx := context.Background()

	opensearchContainer, err := opensearch.Run(
		ctx,
		"opensearchproject/opensearch:2.11.1",
		opensearch.WithUsername("new-username"),
		opensearch.WithPassword("new-password"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := opensearchContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := opensearchContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)
	fmt.Printf("%s : %s\n", opensearchContainer.User, opensearchContainer.Password)

	// Output:
	// true
	// new-username : new-password
}
