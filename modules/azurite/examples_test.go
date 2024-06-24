package azurite_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
)

func ExampleRunContainer() {
	// runAzuriteContainer {
	ctx := context.Background()

	azuriteContainer, err := azurite.RunContainer(
		ctx,
		testcontainers.WithImage("mcr.microsoft.com/azure-storage/azurite:3.28.0"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Clean up the container
	defer func() {
		if err := azuriteContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()
	// }

	state, err := azuriteContainer.State(ctx)
	if err != nil {
		log.Fatalf("failed to get container state: %s", err) // nolint:gocritic
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// This example demonstrates how to create a container, upload a blob, list blobs, and delete the container.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/storage/azblob/examples_test.go#L36
func ExampleRunContainer_blobOperations() {
	// blobOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.RunContainer(
		ctx,
		testcontainers.WithImage("mcr.microsoft.com/azure-storage/azurite:3.28.0"),
		azurite.WithInMemoryPersistence(64),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := azuriteContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// using the built-in shared key credential type
	cred, err := azblob.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Fatalf("failed to create shared key credential: %s", err) // nolint:gocritic
	}

	// create an azblob.Client for the specified storage account that uses the above credentials
	blobServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.BlobService), azurite.AccountName)

	client, err := azblob.NewClientWithSharedKeyCredential(blobServiceURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %s", err) // nolint:gocritic
	}

	// ===== 1. Create a container =====
	containerName := "testcontainer"
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Fatalf("failed to create container: %s", err)
	}

	// ===== 2. Upload and Download a block blob =====
	blobData := "Hello world!"
	blobName := "HelloWorld.txt"

	_, err = client.UploadStream(context.TODO(),
		containerName,
		blobName,
		strings.NewReader(blobData),
		&azblob.UploadStreamOptions{
			Metadata: map[string]*string{"Foo": to.Ptr("Bar")},
			Tags:     map[string]string{"Year": "2022"},
		})
	if err != nil {
		log.Fatalf("failed to upload blob: %s", err)
	}

	// Download the blob's contents and ensure that the download worked properly
	blobDownloadResponse, err := client.DownloadStream(context.TODO(), containerName, blobName, nil)
	if err != nil {
		log.Fatalf("failed to download blob: %s", err) // nolint:gocritic
	}

	// Use the bytes.Buffer object to read the downloaded data.
	// RetryReaderOptions has a lot of in-depth tuning abilities, but for the sake of simplicity, we'll omit those here.
	reader := blobDownloadResponse.Body
	downloadData, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("failed to read downloaded data: %s", err) // nolint:gocritic
	}

	fmt.Println(string(downloadData))

	err = reader.Close()
	if err != nil {
		return
	}

	// ===== 3. List blobs =====
	// List methods returns a pager object which can be used to iterate over the results of a paging operation.
	// To iterate over a page use the NextPage(context.Context) to fetch the next page of results.
	// PageResponse() can be used to iterate over the results of the specific page.
	pager := client.NewListBlobsFlatPager(containerName, nil)
	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to list blobs: %s", err)
		}

		fmt.Println(len(resp.Segment.BlobItems))
	}

	// Delete the blob.
	_, err = client.DeleteBlob(context.TODO(), containerName, blobName, nil)
	if err != nil {
		log.Fatalf("failed to delete blob: %s", err)
	}

	// Delete the container.
	_, err = client.DeleteContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Fatalf("failed to delete container: %s", err)
	}

	// }

	// Output:
	// Hello world!
	// 1
}

// This example demonstrates how to create, list and delete queues.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/storage/azqueue/samples_test.go#L1
func ExampleRunContainer_queueOperations() {
	// queueOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.RunContainer(
		ctx,
		testcontainers.WithImage("mcr.microsoft.com/azure-storage/azurite:3.28.0"),
		azurite.WithInMemoryPersistence(64),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := azuriteContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// using the built-in shared key credential type
	cred, err := azqueue.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Fatalf("failed to create shared key credential: %s", err) // nolint:gocritic
	}

	// create an azqueue.Client for the specified storage account that uses the above credentials
	queueServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.QueueService), azurite.AccountName)

	client, err := azqueue.NewServiceClientWithSharedKeyCredential(queueServiceURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	queueName := "testqueue"

	_, err = client.CreateQueue(context.TODO(), queueName, &azqueue.CreateOptions{
		Metadata: map[string]*string{"hello": to.Ptr("world")},
	})
	if err != nil {
		log.Fatalf("failed to create queue: %s", err)
	}

	pager := client.NewListQueuesPager(&azqueue.ListQueuesOptions{
		Include: azqueue.ListQueuesInclude{Metadata: true},
	})

	// list pre-existing queues
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			log.Fatalf("failed to list queues: %s", err)
		}

		fmt.Println(len(resp.Queues))
		fmt.Println(*resp.Queues[0].Name)
	}

	// delete the queue
	_, err = client.DeleteQueue(context.TODO(), queueName, &azqueue.DeleteOptions{})
	if err != nil {
		log.Fatalf("failed to delete queue: %s", err)
	}

	// }

	// Output:
	// 1
	// testqueue
}

// This example demonstrates how to create, list and delete tables.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/data/aztables/example_test.go#L1
func ExampleRunContainer_tableOperations() {
	// tableOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.RunContainer(
		ctx,
		testcontainers.WithImage("mcr.microsoft.com/azure-storage/azurite:3.28.0"),
		azurite.WithInMemoryPersistence(64),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	defer func() {
		if err := azuriteContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err) // nolint:gocritic
		}
	}()

	// using the built-in shared key credential type
	cred, err := aztables.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Fatalf("failed to create shared key credential: %s", err) // nolint:gocritic
	}

	// create an aztables.Client for the specified storage account that uses the above credentials
	tablesServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.TableService), azurite.AccountName)

	client, err := aztables.NewServiceClientWithSharedKey(tablesServiceURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %s", err)
	}

	tableName := "fromServiceClient"
	// Create a table
	_, err = client.CreateTable(context.TODO(), tableName, nil)
	if err != nil {
		log.Fatalf("failed to create table: %s", err)
	}

	// List tables
	pager := client.NewListTablesPager(nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			log.Fatalf("failed to list tables: %s", err)
		}

		fmt.Println(len(resp.Tables))
		fmt.Println(*resp.Tables[0].Name)
	}

	// Delete a table
	_, err = client.DeleteTable(context.TODO(), tableName, nil)
	if err != nil {
		panic(err)
	}

	// }

	// Output:
	// 1
	// fromServiceClient
}
