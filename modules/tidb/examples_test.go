package tidb_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/tidb"
)

func ExampleRun() {
	// runTiDBContainer {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx,
		"pingcap/tidb:v8.4.0",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := ctr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	ctx := context.Background()

	ctr, err := tidb.Run(ctx,
		"pingcap/tidb:v8.4.0",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connectionString, err := ctr.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Printf("failed to connect to TiDB: %s", err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Printf("failed to ping TiDB: %s", err)
		return
	}

	var result int
	row := db.QueryRow("SELECT 1")
	if err = row.Scan(&result); err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}

	fmt.Println(result)

	// Output:
	// 1
}
