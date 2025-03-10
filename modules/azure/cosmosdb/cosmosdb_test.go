package cosmosdb_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb"
)

func TestCosmosDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := cosmosdb.Run(ctx, "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}
