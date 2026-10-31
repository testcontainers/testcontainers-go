package memcached_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/memcached"
)

func ExampleRun() {
	// runMemcachedContainer {
	ctx := context.Background()

	memcachedContainer, err := memcached.Run(ctx, "memcached:1.6-alpine")
	defer func() {
		if err := testcontainers.TerminateContainer(memcachedContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := memcachedContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
