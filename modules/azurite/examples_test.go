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

func ExampleRun() {
	// runAzuriteContainer {
	ctx := context.Background()

	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.28.0",
	)
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := azuriteContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// Output:
	// true
}

// This example demonstrates how to create a container, upload a blob, list blobs, and delete the container.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/storage/azblob/examples_test.go#L36
func ExampleRun_blobOperations() {
	// blobOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.28.0",
		azurite.WithInMemoryPersistence(64),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// using the built-in shared key credential type
	cred, err := azblob.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Printf("failed to create shared key credential: %s", err)
		return
	}

	// create an azblob.Client for the specified storage account that uses the above credentials
	blobServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.BlobService), azurite.AccountName)

	client, err := azblob.NewClientWithSharedKeyCredential(blobServiceURL, cred, nil)
	if err != nil {
		log.Printf("failed to create client: %s", err)
		return
	}

	// ===== 1. Create a container =====
	containerName := "testcontainer"
	_, err = client.CreateContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Printf("failed to create container: %s", err)
		return
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
		log.Printf("failed to upload blob: %s", err)
		return
	}

	// Download the blob's contents and ensure that the download worked properly
	blobDownloadResponse, err := client.DownloadStream(context.TODO(), containerName, blobName, nil)
	if err != nil {
		log.Printf("failed to download blob: %s", err)
		return
	}

	// Use the bytes.Buffer object to read the downloaded data.
	// RetryReaderOptions has a lot of in-depth tuning abilities, but for the sake of simplicity, we'll omit those here.
	reader := blobDownloadResponse.Body
	downloadData, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("failed to read downloaded data: %s", err)
		return
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
			log.Printf("failed to list blobs: %s", err)
			return
		}

		fmt.Println(len(resp.Segment.BlobItems))
	}

	// Delete the blob.
	_, err = client.DeleteBlob(context.TODO(), containerName, blobName, nil)
	if err != nil {
		log.Printf("failed to delete blob: %s", err)
		return
	}

	// Delete the container.
	_, err = client.DeleteContainer(context.TODO(), containerName, nil)
	if err != nil {
		log.Printf("failed to delete container: %s", err)
		return
	}

	// }

	// Output:
	// Hello world!
	// 1
}

// This example demonstrates how to create, list and delete queues.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/storage/azqueue/samples_test.go#L1
func ExampleRun_queueOperations() {
	// queueOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.28.0",
		azurite.WithInMemoryPersistence(64),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// using the built-in shared key credential type
	cred, err := azqueue.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Printf("failed to create shared key credential: %s", err)
		return
	}

	// create an azqueue.Client for the specified storage account that uses the above credentials
	queueServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.QueueService), azurite.AccountName)

	client, err := azqueue.NewServiceClientWithSharedKeyCredential(queueServiceURL, cred, nil)
	if err != nil {
		log.Printf("failed to create client: %s", err)
		return
	}

	queueName := "testqueue"

	_, err = client.CreateQueue(context.TODO(), queueName, &azqueue.CreateOptions{
		Metadata: map[string]*string{"hello": to.Ptr("world")},
	})
	if err != nil {
		log.Printf("failed to create queue: %s", err)
		return
	}

	pager := client.NewListQueuesPager(&azqueue.ListQueuesOptions{
		Include: azqueue.ListQueuesInclude{Metadata: true},
	})

	// list pre-existing queues
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			log.Printf("failed to list queues: %s", err)
			return
		}

		fmt.Println(len(resp.Queues))
		fmt.Println(*resp.Queues[0].Name)
	}

	// delete the queue
	_, err = client.DeleteQueue(context.TODO(), queueName, &azqueue.DeleteOptions{})
	if err != nil {
		log.Printf("failed to delete queue: %s", err)
		return
	}

	// }

	// Output:
	// 1
	// testqueue
}

// This example demonstrates how to create, list and delete tables.
// Inspired by https://github.com/Azure/azure-sdk-for-go/blob/718000938221915fb2f3c7522d4fd09f7d74cafb/sdk/data/aztables/example_test.go#L1
func ExampleRun_tableOperations() {
	// tableOperations {
	ctx := context.Background()

	azuriteContainer, err := azurite.Run(
		ctx,
		"mcr.microsoft.com/azure-storage/azurite:3.28.0",
		azurite.WithInMemoryPersistence(64),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	// using the built-in shared key credential type
	cred, err := aztables.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	if err != nil {
		log.Printf("failed to create shared key credential: %s", err)
		return
	}

	// create an aztables.Client for the specified storage account that uses the above credentials
	tablesServiceURL := fmt.Sprintf("%s/%s", azuriteContainer.MustServiceURL(ctx, azurite.TableService), azurite.AccountName)

	client, err := aztables.NewServiceClientWithSharedKey(tablesServiceURL, cred, nil)
	if err != nil {
		log.Printf("failed to create client: %s", err)
		return
	}

	tableName := "fromServiceClient"
	// Create a table
	_, err = client.CreateTable(context.TODO(), tableName, nil)
	if err != nil {
		log.Printf("failed to create table: %s", err)
		return
	}

	// List tables
	pager := client.NewListTablesPager(nil)
	for pager.More() {
		resp, err := pager.NextPage(context.Background())
		if err != nil {
			log.Printf("failed to list tables: %s", err)
			return
		}

		fmt.Println(len(resp.Tables))
		fmt.Println(*resp.Tables[0].Name)
	}

	// Delete a table
	_, err = client.DeleteTable(context.TODO(), tableName, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// }

	// Output:
	// 1
	// fromServiceClient
}
