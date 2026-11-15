package examples_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	// Register the MySQL driver, which also supports MariaDB, for database/sql.
	_ "github.com/go-sql-driver/mysql"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func ExampleRun_connect() {
	// connectMariaDB {
	ctx := context.Background()
	ctr, err := mariadb.Run(ctx, "mariadb:12.3.2",
		mariadb.WithScripts(filepath.Join("..", "testdata", "schema.sql")))
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	connectionString, err := ctr.ConnectionString(ctx, "tls=false")
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Printf("failed to connect to MariaDB: %s", err)
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Printf("failed to ping MariaDB: %s", err)
		return
	}
	var name string
	if err := db.QueryRow("SELECT name from profile").Scan(&name); err != nil {
		log.Printf("failed to query profile: %s", err)
		return
	}
	fmt.Println(name)
	// }

	// Output:
	// profile 1
}
