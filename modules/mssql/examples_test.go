package mssql_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
)

func ExampleRunContainer() {
	// runMSSQLServerContainer {
	ctx := context.Background()

	password := "SuperStrong@Passw0rd"

	mssqlContainer, err := mssql.RunContainer(ctx,
		testcontainers.WithImage("mcr.microsoft.com/mssql/server:2022-RTM-GDR1-ubuntu-20.04"),
		mssql.WithAcceptEULA(),
		mssql.WithPassword(password),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mssqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := mssqlContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
