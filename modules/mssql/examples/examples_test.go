package mssql_test

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
)

func ExampleRun_connect() {
	// connectWithSQLClient {
	ctx := context.Background()
	ctr, err := mssql.Run(ctx, "mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04", mssql.WithAcceptEULA())
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		panic(err)
	}

	// Register the sqlserver driver with a blank import of github.com/microsoft/go-mssqldb.
	connectionString, err := ctr.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var connected int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&connected); err != nil {
		panic(err)
	}
	fmt.Println(connected)
	// }

	// Output:
	// 1
}
