package pinecone_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/pinecone"
)

func ExampleRun() {
	ctx := context.Background()

	pineconeContainer, err := pinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(pineconeContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := pineconeContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
