package yugabytedb_test

import (
	"context"
	"fmt"
	"log"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/yugabytedb"
)

func ExampleRun() {
	ctx := context.Background()

	yugabytedbContainer, err := yugabytedb.Run(
		ctx,
		"yugabytedb/yugabyte:2024.1.3.0-b105",
		yugabytedb.WithYCQLKeyspace("custom-keyspace"),
		yugabytedb.WithYCQLUser("custom-user"),
		yugabytedb.WithYSQLDatabaseName("custom-db"),
		yugabytedb.WithYSQLDatabaseUser("custom-user"),
		yugabytedb.WithYSQLDatabasePassword("custom-password"),
	)

	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	defer func() {
		if err := testcontainers.TerminateContainer(yugabytedbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	state, err := yugabytedbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	// Output: true
	fmt.Println(state.Running)
}
