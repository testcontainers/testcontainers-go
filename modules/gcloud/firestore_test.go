package gcloud_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud"
)

type emulatorCreds struct{}

func (ec emulatorCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer owner"}, nil
}

func (ec emulatorCreds) RequireTransportSecurity() bool {
	return false
}

func ExampleRunFirestoreContainer() {
	// runFirestoreContainer {
	ctx := context.Background()

	firestoreContainer, err := gcloud.RunFirestoreContainer(
		ctx,
		testcontainers.WithImage("gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators"),
		gcloud.WithProjectID("firestore-project"),
	)
	if err != nil {
		log.Fatalf("failed to run container: %v", err)
	}

	// Clean up the container
	defer func() {
		if err := firestoreContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %v", err)
		}
	}()
	// }

	// firestoreClient {
	projectID := firestoreContainer.Settings.ProjectID

	conn, err := grpc.Dial(firestoreContainer.URI, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(emulatorCreds{}))
	if err != nil {
		log.Fatalf("failed to dial: %v", err) // nolint:gocritic
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := firestore.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()
	// }

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
	if err != nil {
		log.Fatalf("failed to create document: %v", err)
	}

	docsnap, err := docRef.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get document: %v", err)
	}

	var saved Person
	if err := docsnap.DataTo(&saved); err != nil {
		log.Fatalf("failed to convert data: %v", err)
	}

	fmt.Println(saved.Firstname, saved.Lastname)

	// Output:
	// Ada Lovelace
}
