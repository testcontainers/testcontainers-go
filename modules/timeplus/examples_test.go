package timeplus_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/timeplus"
)

func ExampleRun() {
	// runTimeplusContainer {
	ctx := context.Background()

	timeplusContainer, err := timeplus.Run(ctx, "timeplus/timeplusd:2.3.37")
	defer func() {
		if err := testcontainers.TerminateContainer(timeplusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := timeplusContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleContainer_HTTPEndpoint() {
	ctx := context.Background()

	timeplusContainer, err := timeplus.Run(ctx, "timeplus/timeplusd:2.3.37")
	defer func() {
		if err := testcontainers.TerminateContainer(timeplusContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	httpEndpoint, err := timeplusContainer.HTTPEndpoint(ctx)
	if err != nil {
		log.Printf("failed to get HTTP endpoint: %s", err)
		return
	}

	fmt.Println(len(httpEndpoint) > 0)

	// Output:
	// true
}
