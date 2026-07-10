package sqledge_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/sqledge"
)

func TestAzureSQLEdge(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx, "mcr.microsoft.com/azure-sql-edge:2.0.0")
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

func TestAzureSQLEdgeWithCustomPassword(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx,
		"mcr.microsoft.com/azure-sql-edge:2.0.0",
		sqledge.WithPassword("Strong@Passw0rd"),
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

func TestAzureSQLEdgeWithConnectionStringParameters(t *testing.T) {
	ctx := context.Background()

	ctr, err := sqledge.Run(ctx, "mcr.microsoft.com/azure-sql-edge:2.0.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions — extra params are appended after the defaults
	connectionString, err := ctr.ConnectionString(ctx, "app name=sqledge-test")
	require.NoError(t, err)

	db, err := sql.Open("sqlserver", connectionString)
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)
}
