package cosmosdb_test

// ExampleRun_connect_imports {
import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"github.com/testcontainers/testcontainers-go"
	cosmosdbcontainer "github.com/testcontainers/testcontainers-go/modules/azure/cosmosdb"
	cosmosdbexample "github.com/testcontainers/testcontainers-go/modules/azure/examples/cosmosdb"
)

// }

func ExampleRun_connect() {
	// ExampleRun_connect {
	ctx := context.Background()

	cosmosdbContainer, err := cosmosdbcontainer.Run(ctx, "mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator:vnext-preview")
	defer func() {
		if err := testcontainers.TerminateContainer(cosmosdbContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	connString, err := cosmosdbContainer.ConnectionString(ctx)
	if err != nil {
		log.Printf("failed to get connection string: %s", err)
		return
	}

	p, err := cosmosdbexample.NewContainerPolicy(ctx, cosmosdbContainer)
	if err != nil {
		log.Printf("failed to create policy: %s", err)
		return
	}

	client, err := azcosmos.NewClientFromConnectionString(connString, p.ClientOptions())
	if err != nil {
		log.Printf("failed to create cosmosdb client: %s", err)
		return
	}

	createDatabaseResp, err := client.CreateDatabase(ctx, azcosmos.DatabaseProperties{ID: "myDatabase"}, nil)
	if err != nil {
		log.Printf("failed to create database: %s", err)
		return
	}
	fmt.Println(createDatabaseResp.RawResponse.StatusCode == http.StatusCreated)
	// }

	// Output:
	// true
}
