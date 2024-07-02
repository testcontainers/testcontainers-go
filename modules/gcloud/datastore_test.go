package gcloud_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := datastoreContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
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
		log.Fatalf("failed to create client: %v", err) // nolint:gocritic
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
		log.Fatalf("failed to put data: %v", err)
	}

	saved := Task{}
	err = dsClient.Get(ctx, k, &saved)
	if err != nil {
		log.Fatalf("failed to get data: %v", err)
	}

	fmt.Println(saved.Description)

	// Output:
	// my description
}
