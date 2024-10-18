package meilisearch_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/meilisearch"
)

func ExampleRun() {
	// runMeilisearchContainer {
	ctx := context.Background()

	meiliContainer, err := meilisearch.Run(
		ctx,
		"getmeili/meilisearch:v1.10.3",
		meilisearch.WithMasterKey("my-master-key"),
		meilisearch.WithDumpImport("testdata/movies.dump"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(meiliContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := meiliContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)
	fmt.Printf("%s\n", meiliContainer.MasterKey())

	// Output:
	// true
	// my-master-key
}
