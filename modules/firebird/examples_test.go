package firebird_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/firebird"
)

func ExampleRun() {
	// runFirebirdContainer {
	ctx := context.Background()

	firebirdContainer, err := firebird.Run(ctx,
		"jacobalberty/firebird:v3.0",
		firebird.WithDatabase("test.fdb"),
		firebird.WithUsername("test"),
		firebird.WithPassword("test"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(firebirdContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := firebirdContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
