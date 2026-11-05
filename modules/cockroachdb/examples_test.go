package cockroachdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func ExampleRun() {
	// runCockroachDBContainer {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1")
	defer func() {
		if err := testcontainers.TerminateContainer(cockroachdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}
	fmt.Println(state.Running)

	// Output:
	// true
}
