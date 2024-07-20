package surrealdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/surrealdb"
)

func ExampleRun() {
	// runSurrealDBContainer {
	ctx := context.Background()

	surrealdbContainer, err := surrealdb.Run(ctx, "surrealdb/surrealdb:v1.1.1")
	defer func() {
		if err := testcontainers.TerminateContainer(surrealdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Print(err)
		return
	}
	// }

	state, err := surrealdbContainer.State(ctx)
	if err != nil {
		log.Print(err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
