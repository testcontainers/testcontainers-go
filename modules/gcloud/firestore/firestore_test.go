package firestore_test

import (
	"context"
	"testing"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcfirestore "github.com/testcontainers/testcontainers-go/modules/gcloud/firestore"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcfirestore.WithProjectID("firestore-project"),
	)
	testcontainers.CleanupContainer(t, firestoreContainer)
	require.NoError(t, err)

	projectID := firestoreContainer.ProjectID()

	conn, err := grpc.NewClient(firestoreContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := firestore.NewClient(ctx, projectID, options...)
	require.NoError(t, err)
	defer client.Close()

	users := client.Collection("users")
	docRef := users.Doc("alovelace")

	type Person struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	data := Person{
		Firstname: "Ada",
		Lastname:  "Lovelace",
	}
	_, err = docRef.Create(ctx, data)
	require.NoError(t, err)

	docsnap, err := docRef.Get(ctx)
	require.NoError(t, err)

	var saved Person
	require.NoError(t, docsnap.DataTo(&saved))

	require.Equal(t, "Ada", saved.Firstname)
	require.Equal(t, "Lovelace", saved.Lastname)
}

func TestRunWithDatastore(t *testing.T) {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:513.0.0-emulators",
		tcfirestore.WithProjectID("firestore-project"),
		tcfirestore.WithDatastoreMode(),
	)
	testcontainers.CleanupContainer(t, firestoreContainer)
	require.NoError(t, err)

	projectID := firestoreContainer.ProjectID()

	conn, err := grpc.NewClient(firestoreContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := datastore.NewClient(ctx, projectID, options...)
	require.NoError(t, err)
	defer client.Close()

	userKey := datastore.NameKey("users", "alovelace", nil)

	type Person struct {
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
	}

	data := Person{
		Firstname: "Ada",
		Lastname:  "Lovelace",
	}

	_, err = client.Put(ctx, userKey, &data)
	require.NoError(t, err)

	saved := Person{}
	err = client.Get(ctx, userKey, &saved)
	require.NoError(t, err)

	require.Equal(t, "Ada", saved.Firstname)
	require.Equal(t, "Lovelace", saved.Lastname)
}
