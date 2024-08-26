package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func ExampleRun() {
	// runMySQLContainer {
	ctx := context.Background()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()
	// }

	state, err := mysqlContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRun_connect() {
	ctx := context.Background()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}()

	connectionString, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("failed to connect to MySQL: %s", err) // nolint:gocritic
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping MySQL: %s", err)
	}
	stmt, err := db.Prepare("SELECT @@GLOBAL.tmpdir")
	if err != nil {
		log.Fatalf("failed to prepare statement: %s", err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	tmpDir := ""
	err = row.Scan(&tmpDir)
	if err != nil {
		log.Fatalf("failed to scan row: %s", err)
	}

	fmt.Println(tmpDir)

	// Output:
	// /tmp
}
