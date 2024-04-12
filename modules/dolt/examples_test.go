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

func ExampleRunContainer() {
	// runDoltContainer {
	ctx := context.Background()

	doltContainer, err := dolt.RunContainer(ctx,
		testcontainers.WithImage("dolthub/dolt-sql-server:1.32.4"),
		dolt.WithConfigFile(filepath.Join("testdata", "dolt_1_32_4.cnf")),
		dolt.WithDatabase("foo"),
		dolt.WithUsername("root"),
		dolt.WithPassword("password"),
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		log.Fatalf("failed to run dolt container: %s", err) // nolint:gocritic
	}

	// Clean up the container
	defer func() {
		if err := doltContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate dolt container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := doltContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	ctx := context.Background()

	doltContainer, err := dolt.RunContainer(ctx,
		testcontainers.WithImage("dolthub/dolt-sql-server:1.32.4"),
		dolt.WithConfigFile(filepath.Join("testdata", "dolt_1_32_4.cnf")),
		dolt.WithDatabase("foo"),
		dolt.WithUsername("bar"),
		dolt.WithPassword("password"),
		dolt.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		log.Fatalf("failed to run dolt container: %s", err) // nolint:gocritic
	}

	defer func() {
		if err := doltContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate dolt container: %s", err) // nolint:gocritic
		}
	}()

	connectionString := doltContainer.MustConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("failed to open database connection: %s", err) // nolint:gocritic
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %s", err) // nolint:gocritic
	}
	stmt, err := db.Prepare("SELECT dolt_version();")
	if err != nil {
		log.Fatalf("failed to prepate sql statement: %s", err) // nolint:gocritic
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	version := ""
	err = row.Scan(&version)
	if err != nil {
		log.Fatalf("failed to scan row: %s", err) // nolint:gocritic
	}

	fmt.Println(version)

	// Output:
	// 1.32.4
}
