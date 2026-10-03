package forgejo_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/forgejo"
)

func ExampleRun() {
	// runForgejoContainer {
	ctx := context.Background()

	forgejoContainer, err := forgejo.Run(ctx, "codeberg.org/forgejo/forgejo:11")
	defer func() {
		if err := testcontainers.TerminateContainer(forgejoContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := forgejoContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
