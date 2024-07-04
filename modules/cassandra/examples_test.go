package cassandra_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func ExampleRun() {
	// runCassandraContainer {
	ctx := context.Background()

	cassandraContainer, err := cassandra.Run(ctx,
		"cassandra:4.1.3",
		cassandra.WithInitScripts(filepath.Join("testdata", "init.cql")),
		cassandra.WithConfigFile(filepath.Join("testdata", "config.yaml")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := cassandraContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	connectionHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Fatalf("failed to get connection host: %s", err)
	}

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("failed to create session: %s", err)
	}
	defer session.Close()

	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	if err != nil {
		log.Fatalf("failed to query: %s", err)
	}

	fmt.Println(version)

	// Output:
	// true
	// 4.1.3
}
