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

func ExampleRun() {
	// runClickHouseContainer {
	ctx := context.Background()

	user := "clickhouse"
	password := "password"
	dbname := "testdb"

	clickHouseContainer, err := clickhouse.Run(ctx,
		"clickhouse/clickhouse-server:23.3.8.21-alpine",
		clickhouse.WithUsername(user),
		clickhouse.WithPassword(password),
		clickhouse.WithDatabase(dbname),
		clickhouse.WithInitScripts(filepath.Join("testdata", "init-db.sh")),
		clickhouse.WithConfigFile(filepath.Join("testdata", "config.xml")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(clickHouseContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := clickHouseContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionString, err := clickHouseContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	opts, err := ch.ParseDSN(connectionString)
	if err != nil {
		log.Printf("failed to parse DSN: %s", err)
		return
	}

	fmt.Println(strings.HasPrefix(opts.ClientInfo.String(), "clickhouse-go/"))

	// Output:
	// true
	// true
}
