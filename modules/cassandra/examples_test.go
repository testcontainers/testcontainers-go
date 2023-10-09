package cassandra_test

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func ExampleRunContainer() {
	// runCassandraContainer {
	ctx := context.Background()

	cassandraContainer, err := cassandra.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.1.3"),
		cassandra.WithInitScripts(filepath.Join("testdata", "init.cql")),
		cassandra.WithConfigFile(filepath.Join("testdata", "config.yaml")),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := cassandraContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := cassandraContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	connectionHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		panic(err)
	}

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	if err != nil {
		panic(err)
	}

	fmt.Println(version)

	// Output:
	// true
	// 4.1.3
}
