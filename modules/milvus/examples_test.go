package milvus_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/milvus"
)

func ExampleRun() {
	// runMilvusContainer {
	ctx := context.Background()

	milvusContainer, err := milvus.Run(ctx, "milvusdb/milvus:v2.6.3")
	defer func() {
		if err := testcontainers.TerminateContainer(milvusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := milvusContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
