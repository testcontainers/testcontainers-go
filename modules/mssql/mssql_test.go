package mssql_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mssql"
	"github.com/testcontainers/testcontainers-go/wait"
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

func TestMSSQLServerWithMissingEulaOption(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		testcontainers.WithWaitStrategy(
			wait.ForLog("The SQL Server End-User License Agreement (EULA) must be accepted")),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	state, err := ctr.State(ctx)
	require.NoError(t, err)

	if !state.Running {
		t.Log("Success: Confirmed proper handling of missing EULA, so container is not running.")
	}
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

// tests that a weak password is not accepted by the container due to Microsoft's password strength policy
func TestMSSQLServerWithInvalidPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := mssql.Run(ctx,
		"mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Password validation failed")),
		mssql.WithAcceptEULA(),
		mssql.WithPassword("weakPassword"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}
