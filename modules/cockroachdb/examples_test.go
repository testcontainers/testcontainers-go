package cockroachdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cockroachdb"
)

func ExampleRun() {
	// runCockroachDBContainer {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1")
	defer func() {
		if err := testcontainers.TerminateContainer(cockroachdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}
	fmt.Println(state.Running)

	cfg, err := cockroachdbContainer.ConnectionConfig(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		log.Printf("failed to connect: %s", err)
		return
	}

	defer func() {
		if err := conn.Close(ctx); err != nil {
			log.Printf("failed to close connection: %s", err)
		}
	}()

	if err = conn.Ping(ctx); err != nil {
		log.Printf("failed to ping: %s", err)
		return
	}

	// Output:
	// true
}

func ExampleRun_withInitOptions() {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1",
		cockroachdb.WithNoClusterDefaults(),
		cockroachdb.WithInitScripts("testdata/__init.sql"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cockroachdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := cockroachdbContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}
	fmt.Println(state.Running)

	addr, err := cockroachdbContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("pgx/v5", addr)
	if err != nil {
		log.Printf("failed to open connection: %s", err)
		return
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close connection: %s", err)
		}
	}()

	var interval string
	if err := db.QueryRow("SHOW CLUSTER SETTING kv.range_merge.queue_interval").Scan(&interval); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}
	fmt.Println(interval)

	if err := db.QueryRow("SHOW CLUSTER SETTING jobs.registry.interval.gc").Scan(&interval); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}
	fmt.Println(interval)

	var statsCollectionEnabled bool
	if err := db.QueryRow("SHOW CLUSTER SETTING sql.stats.automatic_collection.enabled").Scan(&statsCollectionEnabled); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}
	fmt.Println(statsCollectionEnabled)

	// Output:
	// true
	// 00:00:05
	// 00:00:50
	// true
}
