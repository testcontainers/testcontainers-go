package atlaslocal_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/atlaslocal"
)

func ExampleRun() {
	ctx := context.Background()

	atlaslocalContainer, err := atlaslocal.Run(ctx, "mongodb/mongodb-atlas-local:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(atlaslocalContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := atlaslocalContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
