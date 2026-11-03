package arangodb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	tcarangodb "github.com/testcontainers/testcontainers-go/modules/arangodb"
)

func ExampleRun() {
	// runArangoDBContainer {
	ctx := context.Background()

	const password = "t3stc0ntain3rs!"

	arangodbContainer, err := tcarangodb.Run(ctx, "arangodb:3.11.5", tcarangodb.WithRootPassword(password))
	defer func() {
		if err := testcontainers.TerminateContainer(arangodbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := arangodbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
