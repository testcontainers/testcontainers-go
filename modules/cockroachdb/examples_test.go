package cockroachdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"

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

	addr, err := cockroachdbContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}
	u, err := url.Parse(addr)
	if err != nil {
		log.Printf("failed to parse connection string: %s", err)
		return
	}
	u.Host = fmt.Sprintf("%s:%s", u.Hostname(), "xxx")
	fmt.Println(u.String())

	// Output:
	// true
	// postgres://root@localhost:xxx/defaultdb?sslmode=disable
}

func ExampleRun_withRecommendedSettings() {
	ctx := context.Background()

	cockroachdbContainer, err := cockroachdb.Run(ctx, "cockroachdb/cockroach:latest-v23.1", cockroachdb.WithStatements(cockroachdb.ClusterDefaults...))
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

	var queueInterval string
	if err := db.QueryRow("SHOW CLUSTER SETTING kv.range_merge.queue_interval").Scan(&queueInterval); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}
	fmt.Println(queueInterval)

	var statsCollectionEnabled bool
	if err := db.QueryRow("SHOW CLUSTER SETTING sql.stats.automatic_collection.enabled").Scan(&statsCollectionEnabled); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}
	fmt.Println(statsCollectionEnabled)

	// Output:
	// true
	// 00:00:00.05
	// false
}
