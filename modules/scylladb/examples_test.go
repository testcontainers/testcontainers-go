package scylladb_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gocql/gocql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/scylladb"
)

func ExampleRun_withCustomCommands() {
	ctx := context.Background()

	// runScyllaDBContainerWithCustomCommands {
	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithCustomCommands("--memory=1G", "--smp=2"),
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

	connectionHost, err := scyllaContainer.NonShardAwareConnectionHost(ctx)
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	if err := runGoCQLExampleTest(connectionHost); err != nil {
		log.Printf("failed to run Go CQL example test: %s", err)
		return
	}

	// Output:
	// true
}

func ExampleRun_withAlternator() {
	ctx := context.Background()

	// runScyllaDBContainerWithAlternator {

	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithAlternator(),
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

	// scyllaDbAlternatorConnectionHost {
	// the alternator port is the default port 8000
	_, err = scyllaContainer.AlternatorConnectionHost(ctx)
	// }
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

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

	// scyllaDbShardAwareConnectionHost {
	connectionHost, err := scyllaContainer.ShardAwareConnectionHost(ctx)
	// }
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	if err := runGoCQLExampleTest(connectionHost); err != nil {
		log.Printf("failed to run Go CQL example test: %s", err)
		return
	}

	// Output:
	// true
}

func ExampleRun_withConfig() {
	// runScyllaDBContainerWithConfig {
	ctx := context.Background()

	cfgBytes := `cluster_name: 'Amazing ScyllaDB Test'
num_tokens: 256
commitlog_sync: periodic
commitlog_sync_period_in_ms: 10000
commitlog_segment_size_in_mb: 32
schema_commitlog_segment_size_in_mb: 128
seed_provider:
  - class_name: org.apache.cassandra.locator.SimpleSeedProvider
    parameters:
      - seeds: "127.0.0.1"
listen_address: localhost
native_transport_port: 9042
native_shard_aware_transport_port: 19042
read_request_timeout_in_ms: 5000
write_request_timeout_in_ms: 2000
cas_contention_timeout_in_ms: 1000
endpoint_snitch: SimpleSnitch
rpc_address: localhost
api_port: 10000
api_address: 127.0.0.1
batch_size_warn_threshold_in_kb: 128
batch_size_fail_threshold_in_kb: 1024
partitioner: org.apache.cassandra.dht.Murmur3Partitioner
commitlog_total_space_in_mb: -1
murmur3_partitioner_ignore_msb_bits: 12
strict_is_not_null_in_views: true
maintenance_socket: ignore
enable_tablets: true
`

	scyllaContainer, err := scylladb.Run(ctx,
		"scylladb/scylla:6.2",
		scylladb.WithConfig(strings.NewReader(cfgBytes)),
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

	connectionHost, err := scyllaContainer.NonShardAwareConnectionHost(ctx)
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	if err := runGoCQLExampleTest(connectionHost); err != nil {
		log.Printf("failed to run Go CQL example test: %s", err)
		return
	}

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

	// scyllaDbNonShardAwareConnectionHost {
	connectionHost, err := scyllaContainer.NonShardAwareConnectionHost(ctx)
	// }
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	if err := runGoCQLExampleTest(connectionHost); err != nil {
		log.Printf("failed to run Go CQL example test: %s", err)
		return
	}

	// Output:
	// true
}

func runGoCQLExampleTest(connectionHost string) error {
	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("create cluster session: %w", err)
	}
	defer session.Close()

	var driver string
	err = session.Query("SELECT driver_name FROM system.clients").Scan(&driver)
	if err != nil {
		return fmt.Errorf("session query: %w", err)
	}

	return nil
}
