package inbucket_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/inbucket"
)

func ExampleRun() {
	// runInbucketContainer {
	ctx := context.Background()

	inbucketContainer, err := inbucket.Run(ctx, "inbucket/inbucket:sha-2d409bb")
	defer func() {
		if err := testcontainers.TerminateContainer(inbucketContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := inbucketContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
