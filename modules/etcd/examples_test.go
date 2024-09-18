package etcd_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
)

func ExampleRun() {
	// runetcdContainer {
	ctx := context.Background()

	etcdContainer, err := etcd.Run(ctx, "bitnami/etcd:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(etcdContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := etcdContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
