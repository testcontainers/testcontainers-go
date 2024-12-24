package dolt_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dolt"
)

func ExampleRun() {
	// runDoltContainer {
	ctx := context.Background()

	doltContainer, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithConfigFile(filepath.Join("testdata", "dolt_1_32_4.cnf")),
		dolt.WithDatabase("foo"),
		dolt.WithUsername("root"),
		dolt.WithPassword("password"),
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(doltContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run dolt container: %s", err)
		return
	}
	// }

	state, err := doltContainer.State(ctx)
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

	doltContainer, err := dolt.Run(ctx,
		"dolthub/dolt-sql-server:1.32.4",
		dolt.WithConfigFile(filepath.Join("testdata", "dolt_1_32_4.cnf")),
		dolt.WithDatabase("foo"),
		dolt.WithUsername("bar"),
		dolt.WithPassword("password"),
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(doltContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run dolt container: %s", err)
		return
	}

	connectionString := doltContainer.MustConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Printf("failed to open database connection: %s", err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Printf("failed to ping database: %s", err)
		return
	}
	stmt, err := db.Prepare("SELECT dolt_version();")
	if err != nil {
		log.Printf("failed to prepate sql statement: %s", err)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	version := ""
	err = row.Scan(&version)
	if err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}

	fmt.Println(version)

	// Output:
	// 1.32.4
}
