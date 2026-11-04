package cosmosdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb"
)

func TestCosmosDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := cosmosdb.Run(ctx, "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Get the connection string.
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotNil(t, connStr)
}
