package yugabytedb_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/yugabytedb"
	"github.com/yugabyte/gocql"
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

func ExampleYugabyteDBContainer_YSQLConnectionString() {
	ctx := context.Background()

	yugabytedbContainer, err := yugabytedb.Run(
		ctx,
		"yugabytedb/yugabyte:2024.1.3.0-b105",
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

	connStr, err := yugabytedbContainer.YSQLConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("failed to open connection: %s", err)
		return
	}

	defer db.Close()
}

func ExampleYugabyteDBContainer_YCQLConfigureClusterConfig() {
	ctx := context.Background()

	yugabytedbContainer, err := yugabytedb.Run(
		ctx,
		"yugabytedb/yugabyte:2024.1.3.0-b105",
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

	cluster := gocql.NewCluster()
	yugabytedbContainer.YCQLConfigureClusterConfig(ctx, cluster)

	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}

	defer session.Close()
}
