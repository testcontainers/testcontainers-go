package cosmosdb_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
	"github.com/stretchr/testify/require"
	
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb"
)

func TestCosmosDB(t *testing.T) {
	ctx := context.Background()

	ctr, err := cosmosdb.Run(ctx, "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Create Azure Cosmos client
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotNil(t, connStr)

	p, err := cosmosdb.NewContainerPolicy(ctx, ctr)
	require.NoError(t, err)

	client, err := azcosmos.NewClientFromConnectionString(connStr, p.ClientOptions())
	require.NoError(t, err)
	require.NotNil(t, client)

	// Create database
	createDatabaseResp, err := client.CreateDatabase(ctx, azcosmos.DatabaseProperties{ID: "myDatabase"}, nil)
	require.NoError(t, err)
	require.NotNil(t, createDatabaseResp)

	dbClient, err := client.NewDatabase("myDatabase")
	require.NoError(t, err)
	require.NotNil(t, dbClient)

	// Create container
	containerProps := azcosmos.ContainerProperties{
		ID:                     "myContainer",
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{Paths: []string{"/category"}},
	}
	createContainerResp, err := dbClient.CreateContainer(ctx, containerProps, nil)
	require.NoError(t, err)
	require.NotNil(t, createContainerResp)
	containerClient, err := dbClient.NewContainer("myContainer")
	require.NoError(t, err)
	require.NotNil(t, containerClient)

	// Create item
	type Product struct {
		ID       string `json:"id"`
		Category string `json:"category"`
		Name     string `json:"name"`
	}

	testItem := Product{ID: "item123", Category: "gear-surf-surfboards", Name: "Yamba Surfboard"}

	pk := azcosmos.NewPartitionKeyString(testItem.Category)

	jsonItem, err := json.Marshal(testItem)
	require.NoError(t, err)

	createItemResp, err := containerClient.CreateItem(ctx, pk, jsonItem, nil)
	require.NoError(t, err)
	require.NotNil(t, createItemResp)

	// Read item
	readItemResp, err := containerClient.ReadItem(ctx, pk, testItem.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, readItemResp)
}
