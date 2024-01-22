package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func ExampleRunContainer() {
	// runMySQLContainer {
	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0.36"),
		mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		panic(err)
	}

	// Clean up the container
	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()
	// }

	state, err := mysqlContainer.State(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

func ExampleRunContainer_connect() {
	ctx := context.Background()

	mysqlContainer, err := mysql.RunContainer(ctx,
		testcontainers.WithImage("mysql:8.0.36"),
		mysql.WithConfigFile(filepath.Join("testdata", "my_8.cnf")),
		mysql.WithDatabase("foo"),
		mysql.WithUsername("root"),
		mysql.WithPassword("password"),
		mysql.WithScripts(filepath.Join("testdata", "schema.sql")),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := mysqlContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	connectionString, _ := mysqlContainer.ConnectionString(ctx)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		panic(err)
	}
	stmt, err := db.Prepare("SELECT @@GLOBAL.tmpdir")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	tmpDir := ""
	err = row.Scan(&tmpDir)
	if err != nil {
		panic(err)
	}

	fmt.Println(tmpDir)

	// Output:
	// /tmp
}
