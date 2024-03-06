package surrealdb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/surrealdb"
)

func ExampleRunContainer() {
	// runSurrealDBContainer {
	ctx := context.Background()

	surrealdbContainer, err := surrealdb.RunContainer(ctx, testcontainers.WithImage("surrealdb/surrealdb:v1.1.1"))
	if err != nil {
		log.Fatal(err)
	}

	// Clean up the container
	defer func() {
		if err := surrealdbContainer.Terminate(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	// }

	state, err := surrealdbContainer.State(ctx)
	if err != nil {
		log.Fatal(err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
