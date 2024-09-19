package mssql_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
)

func ExampleRun() {
	// runMSSQLServerContainer {
	ctx := context.Background()

	password := "SuperStrong@Passw0rd"

	mssqlContainer, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword(password),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(mssqlContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mssqlContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}
