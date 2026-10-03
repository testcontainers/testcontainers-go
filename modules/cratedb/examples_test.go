package cratedb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cratedb"
)

func ExampleRun() {
	// runCrateDBContainer {
	ctx := context.Background()

	cratedbContainer, err := cratedb.Run(ctx,
		"crate:5.7",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cratedbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cratedbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
