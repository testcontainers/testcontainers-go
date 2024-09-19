package opensearch_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(opensearchContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := opensearchContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)
	fmt.Printf("%s : %s\n", opensearchContainer.User, opensearchContainer.Password)

	// Output:
	// true
	// new-username : new-password
}
