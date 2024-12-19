package scylladb_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/scylladb"
)

func ExampleRun_withCustomCommands() {
	ctx := context.Background()

	// runScyllaDBContainerWithCustomCommands {
	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithCustomCommand("--memory", "1G"),
		scylladb.WithCustomCommand("--smp", "2"),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(scyllaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := scyllaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := scyllaContainer.ConnectionHost(ctx, 9042) // Non ShardAwareness port
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	runGoCQLExampleTest(connectionHost)

	// Output:
	// true
}

func ExampleRun_withAlternator() {
	ctx := context.Background()

	// runScyllaDBContainerWithAlternator {
	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithAlternator(8000), // Choose which port to use on Alternator
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(scyllaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := scyllaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := scyllaContainer.ConnectionHost(ctx, 8080) // Alternator port
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	runGoCQLExampleTest(connectionHost)

	// Output:
	// true
}

func ExampleRun_withShardAwareness() {
	ctx := context.Background()

	// runScyllaDBContainerWithShardAwareness {
	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithShardAwareness(),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(scyllaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := scyllaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := scyllaContainer.ConnectionHost(ctx, 19042) // ShardAwareness port
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	runGoCQLExampleTest(connectionHost)

	// Output:
	// true
}

func ExampleRun_withConfigFile() {
	ctx := context.Background()

	// runScyllaDBContainerWithConfigFile {
	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithConfigFile(filepath.Join("testdata", "scylla.yaml")),
	)
	// }
	defer func() {
		if err := testcontainers.TerminateContainer(scyllaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := scyllaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := scyllaContainer.ConnectionHost(ctx, 9042) // Non ShardAwareness port
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	runGoCQLExampleTest(connectionHost)

	// Output:
	// true
}

func ExampleRun() {
	// runBaseScyllaDBContainer {
	ctx := context.Background()

	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithShardAwareness(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(scyllaContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := scyllaContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// BaseConnectionHost {
	connectionHost, err := scyllaContainer.ConnectionHost(ctx, 9042)
	// }
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}
	runGoCQLExampleTest(connectionHost)

	// Output:
	// true
}

func runGoCQLExampleTest(connectionHost string) {
	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
	}
	defer session.Close()

	var driver string
	err = session.Query("SELECT driver_name FROM system.clients").Scan(&driver)
	if err != nil {
		log.Printf("failed to query: %s", err)
	}
}
