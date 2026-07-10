package orientdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/orientdb"
)

func ExampleRun() {
	ctx := context.Background()

	orientdbContainer, err := orientdb.Run(ctx, "orientdb:3.2",
		orientdb.WithRootPassword("rootpwd"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(orientdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := orientdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
