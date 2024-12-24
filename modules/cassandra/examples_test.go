package cassandra_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(cassandraContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cassandraContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}
	defer session.Close()

	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	if err != nil {
		log.Printf("failed to query: %s", err)
		return
	}

	fmt.Println(version)

	// Output:
	// true
	// 4.1.3
}
