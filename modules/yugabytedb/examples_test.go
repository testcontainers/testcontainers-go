package yugabytedb_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	_ "github.com/lib/pq"
	"github.com/yugabyte/gocql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/yugabytedb"
)

func ExampleRun() {
	// runyugabyteDBContainer {
	ctx := context.Background()

	yugabytedbContainer, err := yugabytedb.Run(
		ctx,
		"yugabytedb/yugabyte:2024.1.3.0-b105",
		yugabytedb.WithKeyspace("custom-keyspace"),
		yugabytedb.WithUser("custom-user"),
		yugabytedb.WithDatabaseName("custom-db"),
		yugabytedb.WithDatabaseUser("custom-user"),
		yugabytedb.WithDatabasePassword("custom-password"),
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
	// }

	state, err := yugabytedbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output: true
}

func ExampleContainer_YSQLConnectionString() {
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

	var i int
	row := db.QueryRowContext(ctx, "SELECT 1")
	if err := row.Scan(&i); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}

	fmt.Println(i)

	// Output: 1
}

func ExampleContainer_newCluster() {
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

	yugabytedbContainerHost, err := yugabytedbContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get container host: %s", err)
		return
	}

	yugabyteContainerPort, err := yugabytedbContainer.MappedPort(ctx, "9042/tcp")
	if err != nil {
		log.Printf("failed to get container port: %s", err)
		return
	}

	cluster := gocql.NewCluster(net.JoinHostPort(yugabytedbContainerHost, yugabyteContainerPort.Port()))
	cluster.Keyspace = "yugabyte"
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "yugabyte",
		Password: "yugabyte",
	}

	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}

	defer session.Close()

	var i int
	if err := session.Query(`
		SELECT COUNT(*) 
		FROM system_schema.keyspaces 
		WHERE keyspace_name = 'yugabyte'
	`).Scan(&i); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}

	fmt.Println(i)

	// Output: 1
}
