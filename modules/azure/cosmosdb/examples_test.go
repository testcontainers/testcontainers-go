package cosmosdb_test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb"
)

func ExampleRun() {
	// runCosmosDBContainer {
	ctx := context.Background()

	cosmosdbCtr, err := cosmosdb.Run(
		ctx,
		"mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cosmosdbCtr); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cosmosdbCtr.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// ExampleRun_authenticateCreateClient is inspired by the example from the Azure CosmosDB Go SDK:
// https://docs.azure.cn/en-us/cosmos-db/nosql/samples-go
func ExampleRun_authenticateCreateClient() {
	// ===== 1. Run the CosmosDB container =====
	ctx := context.Background()

	cosmosdbContainer, err := cosmosdb.Run(
		ctx,
		"mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:latest",
		cosmosdb.WithPartitionCount(3),
		cosmosdb.WithLogLevel("debug"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cosmosdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// ===== 2. Create a CosmosDB client using a connection string to the container =====
	// createClient {
	clientOpts := azcosmos.ClientOptions{
		ClientOptions: policy.ClientOptions{
			Transport: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			},
		},
	}

	client, err := azcosmos.NewClientFromConnectionString(cosmosdbContainer.MustConnectionString(ctx), &clientOpts)
	if err != nil {
		log.Printf("failed to create client: %s", err)
		return
	}
	// }

	// ===== 3. Create a CosmosDB database =====
	// createDatabase {
	dbName := "myDatabase"

	databaseClient, err := client.NewDatabase(dbName)
	if err != nil {
		log.Printf("failed to create database: %s", err)
		return
	}
	// }

	// ===== 4. Create a CosmosDB container =====
	// createContainer {
	containerName := "myContainer"

	cosmosDBContainerClient, err := databaseClient.NewContainer(containerName)
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return
	}
	// }

	fmt.Println(cosmosDBContainerClient.ID())

	// ===== 5. CRUD Operations =====
	item := map[string]string{
		"id":    "1",
		"value": "2",
	}

	marshalled, err := json.Marshal(item)
	if err != nil {
		log.Printf("failed to marshal item: %s", err)
		return
	}

	pk := azcosmos.NewPartitionKeyString("1")
	id := "1"

	// Create an item
	itemResponse, err := cosmosDBContainerClient.CreateItem(context.Background(), pk, marshalled, nil)
	if err != nil {
		log.Printf("failed to create item: %s", err)
		return
	}

	// Read an item
	itemResponse, err = cosmosDBContainerClient.ReadItem(context.Background(), pk, id, nil)
	if err != nil {
		log.Printf("failed to read item: %s", err)
		return
	}

	var itemResponseBody map[string]string
	err = json.Unmarshal(itemResponse.Value, &itemResponseBody)
	if err != nil {
		log.Printf("failed to unmarshal item: %s", err)
		return
	}

	itemResponseBody["value"] = "3"
	marshalledReplace, err := json.Marshal(itemResponseBody)
	if err != nil {
		log.Printf("failed to marshal item: %s", err)
		return
	}

	// Replace an item
	itemResponse, err = cosmosDBContainerClient.ReplaceItem(context.Background(), pk, id, marshalledReplace, nil)
	if err != nil {
		log.Printf("failed to replace item: %s", err)
		return
	}

	// Patch an item
	patch := azcosmos.PatchOperations{}

	patch.AppendAdd("/newField", "newValue")
	patch.AppendRemove("/oldFieldToRemove")

	itemResponse, err = cosmosDBContainerClient.PatchItem(context.Background(), pk, id, patch, nil)
	if err != nil {
		log.Printf("failed to patch item: %s", err)
		return
	}

	// Delete an item
	itemResponse, err = cosmosDBContainerClient.DeleteItem(context.Background(), pk, id, nil)
	if err != nil {
		log.Printf("failed to delete item: %s", err)
		return
	}

	// Output:
	// myDatabase
	// myContainer
}
