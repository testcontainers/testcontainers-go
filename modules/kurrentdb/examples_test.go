package kurrentdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kurrentdb"
)

func ExampleRun() {
	// runKurrentDBContainer {
	ctx := context.Background()

	kurrentdbContainer, err := kurrentdb.Run(ctx, "kurrentplatform/kurrentdb:latest")
	defer func() {
		if err := testcontainers.TerminateContainer(kurrentdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := kurrentdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_withInsecure() {
	ctx := context.Background()

	// withInsecure {
	kurrentdbContainer, err := kurrentdb.Run(ctx,
		"kurrentplatform/kurrentdb:latest",
		kurrentdb.WithInsecure(),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(kurrentdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// connectionString {
	connStr, err := kurrentdbContainer.ConnectionString(ctx)
	// }
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	fmt.Println(connStr != "")

	// Output:
	// true
}
