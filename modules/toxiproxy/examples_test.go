package toxiproxy_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/toxiproxy"
)

func ExampleRun() {
	ctx := context.Background()

	toxiproxyContainer, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
	defer func() {
		if err := testcontainers.TerminateContainer(toxiproxyContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := toxiproxyContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
