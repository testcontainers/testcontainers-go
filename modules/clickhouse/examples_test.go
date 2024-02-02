package clickhouse_test

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
)

func ExampleRunContainer() {
	// runClickHouseContainer {
	ctx := context.Background()

	user := "clickhouse"
	password := "password"
	dbname := "testdb"

	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:23.3.8.21-alpine"),
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
		clickhouse.WithInitScripts(filepath.Join("testdata", "init-db.sh")),
		clickhouse.WithConfigFile(filepath.Join("testdata", "config.xml")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}
	defer func() {
		if err := clickHouseContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := clickHouseContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	connectionString, err := clickHouseContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	opts, err := ch.ParseDSN(connectionString)
	if err != nil {
		log.Fatalf("failed to parse DSN: %s", err)
	}

	fmt.Println(strings.HasPrefix(opts.ClientInfo.String(), "clickhouse-go/"))

	// Output:
	// true
	// true
}
