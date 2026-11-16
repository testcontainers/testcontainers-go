package mssql_test

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"net/url"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
)

func TestMSSQLServer(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("sqlserver", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE a_table ( " +
		" [col_1] NVARCHAR(128) NOT NULL, " +
		" [col_2] NVARCHAR(128) NOT NULL, " +
		" PRIMARY KEY ([col_1], [col_2]) " +
		")")
	require.NoError(t, err)
}

func TestMSSQLServerWithConnectionStringParameters(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
	require.NoError(t, err)

	parsed, err := url.Parse(connectionString)
	require.NoError(t, err)
	require.Equal(t, "false", parsed.Query().Get("encrypt"))
	require.Equal(t, "true", parsed.Query().Get("TrustServerCertificate"))

	db, err := sql.Open("sqlserver", connectionString)
	require.NoError(t, err)

	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE a_table ( " +
		" [col_1] NVARCHAR(128) NOT NULL, " +
		" [col_2] NVARCHAR(128) NOT NULL, " +
		" PRIMARY KEY ([col_1], [col_2]) " +
		")")
	require.NoError(t, err)
}

func TestMSSQLServerWithCustomStrongPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		mssql.WithAcceptEULA(),
		mssql.WithPassword("Strong@Passw0rd"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
	connectionString, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	db, err := sql.Open("sqlserver", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)
}

//go:embed testdata/seed.sql
var seedSQLContent []byte

// tests that a container can be created with a DDL script
func TestMSSQLServerWithScriptsDDL(t *testing.T) {
	const password = "MyCustom@Passw0rd"

	// assertContainer contains the logic for asserting the test
	assertContainer := func(t *testing.T, ctx context.Context, image string, options ...testcontainers.ContainerCustomizer) {
		t.Helper()

		ctr, err := mssql.Run(ctx,
			image,
			append([]testcontainers.ContainerCustomizer{mssql.WithAcceptEULA()}, options...)...,
		)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		connectionString, err := ctr.ConnectionString(ctx)
		require.NoError(t, err)

		db, err := sql.Open("sqlserver", connectionString)
		require.NoError(t, err)
		defer db.Close()

		err = db.PingContext(ctx)
		require.NoError(t, err)

		rows, err := db.QueryContext(ctx, "SELECT * FROM pizza_palace.pizzas")
		require.NoError(t, err)
		defer rows.Close()

		type Pizza struct {
			ID            int
			ToppingName   string
			Deliciousness string
		}

		want := []Pizza{
			{1, "Pineapple", "Controversial but tasty"},
			{2, "Pepperoni", "Classic never fails"},
		}
		got := make([]Pizza, 0, len(want))

		for rows.Next() {
			var p Pizza
			err := rows.Scan(&p.ID, &p.ToppingName, &p.Deliciousness)
			require.NoError(t, err)
			got = append(got, p)
		}

		require.NoError(t, rows.Err())
		require.Equal(t, want, got)
	}

	ctx := context.Background()

	t.Run("WithPassword/beforeWithScripts", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			mssql.WithPassword(password),
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
		)
	})

	t.Run("WithPassword/afterWithScripts", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
			mssql.WithPassword(password),
		)
	})

	t.Run("2019-CU30-ubuntu-20.04/oldSQLCmd", func(t *testing.T) {
		assertContainer(t, ctx,
			"mcr.microsoft.com/mssql/server:2019-CU30-ubuntu-20.04",
			mssql.WithPassword(password),
			mssql.WithInitSQL(bytes.NewReader(seedSQLContent)),
		)
	})
}
