package datastore_test

import (
	"context"
	"log"
	"testing"

	"cloud.google.com/go/datastore"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcdatastore "github.com/testcontainers/testcontainers-go/modules/gcloud/datastore"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	datastoreContainer, err := tcdatastore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcdatastore.WithProjectID("datastore-project"),
	)
	testcontainers.CleanupContainer(t, datastoreContainer)
	require.NoError(t, err)

	projectID := datastoreContainer.ProjectID()

	options := []option.ClientOption{
		option.WithEndpoint(datastoreContainer.URI()),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}

	dsClient, err := datastore.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
	}
	defer dsClient.Close()

	type Task struct {
		Description string
	}

	k := datastore.NameKey("Task", "sample", nil)
	data := Task{
		Description: "my description",
	}
	_, err = dsClient.Put(ctx, k, &data)
	require.NoError(t, err)

	saved := Task{}
	err = dsClient.Get(ctx, k, &saved)
	require.NoError(t, err)

	require.Equal(t, "my description", saved.Description)
}
