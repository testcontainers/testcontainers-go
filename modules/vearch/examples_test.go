package vearch_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vearch"
)

func ExampleRun() {
	// runVearchContainer {
	ctx := context.Background()

	vearchContainer, err := vearch.Run(ctx, "vearch/vearch:3.5.1")
	defer func() {
		if err := testcontainers.TerminateContainer(vearchContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := vearchContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
