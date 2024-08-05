package gcloud_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

func ExampleRunDatastoreContainer() {
	// runDatastoreContainer {
	ctx := context.Background()

	datastoreContainer, err := gcloud.RunDatastore(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		gcloud.WithProjectID("datastore-project"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(datastoreContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run container: %v", err)
		return
	}
	// }

	// datastoreClient {
	projectID := datastoreContainer.Settings.ProjectID

	options := []option.ClientOption{
		option.WithEndpoint(datastoreContainer.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}

	dsClient, err := datastore.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
	}
	defer dsClient.Close()
	// }

	type Task struct {
		Description string
	}

	k := datastore.NameKey("Task", "sample", nil)
	data := Task{
		Description: "my description",
	}
	_, err = dsClient.Put(ctx, k, &data)
	if err != nil {
		log.Printf("failed to put data: %v", err)
		return
	}

	saved := Task{}
	err = dsClient.Get(ctx, k, &saved)
	if err != nil {
		log.Printf("failed to get data: %v", err)
		return
	}

	fmt.Println(saved.Description)

	// Output:
	// my description
}
