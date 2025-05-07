package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
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
	defer func() {
		if err := testcontainers.TerminateContainer(mysqlContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := mysqlContainer.State(ctx)
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

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(mysqlContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connectionString, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Printf("failed to connect to MySQL: %s", err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Printf("failed to ping MySQL: %s", err)
		return
	}
	stmt, err := db.Prepare("SELECT @@GLOBAL.tmpdir")
	if err != nil {
		log.Printf("failed to prepare statement: %s", err)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	tmpDir := ""
	err = row.Scan(&tmpDir)
	if err != nil {
		log.Printf("failed to scan row: %s", err)
		return
	}

	fmt.Println(tmpDir)

	// Output:
	// /tmp
}
