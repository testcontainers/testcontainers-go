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

	ctr, err := sqledge.Run(ctx, "mcr.microsoft.com/azure-sql-edge:1.0.7")
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
		"mcr.microsoft.com/azure-sql-edge:1.0.7",
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

	ctr, err := sqledge.Run(ctx, "mcr.microsoft.com/azure-sql-edge:1.0.7")
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
}
