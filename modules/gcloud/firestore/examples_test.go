package firestore_test

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/testcontainers/testcontainers-go"
	tcfirestore "github.com/testcontainers/testcontainers-go/modules/gcloud/firestore"
)

type emulatorCreds struct{}

func (ec emulatorCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer owner"}, nil
}

func (ec emulatorCreds) RequireTransportSecurity() bool {
	return false
}

func ExampleRun() {
	// runFirestoreContainer {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:367.0.0-emulators",
		tcfirestore.WithProjectID("firestore-project"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(firestoreContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run container: %v", err)
		return
	}
	// }

	// firestoreClient {
	projectID := firestoreContainer.ProjectID()

	conn, err := grpc.NewClient(firestoreContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(emulatorCreds{}))
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := firestore.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
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
		log.Printf("failed to create document: %v", err)
		return
	}

	docsnap, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("failed to get document: %v", err)
		return
	}

	var saved Person
	if err := docsnap.DataTo(&saved); err != nil {
		log.Printf("failed to convert data: %v", err)
		return
	}

	fmt.Println(saved.Firstname, saved.Lastname)

	// Output:
	// Ada Lovelace
}

func ExampleRun_datastoreMode() {
	ctx := context.Background()

	firestoreContainer, err := tcfirestore.Run(
		ctx,
		"gcr.io/google.com/cloudsdktool/cloud-sdk:513.0.0-emulators",
		tcfirestore.WithProjectID("firestore-project"),
		tcfirestore.WithDatastoreMode(),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(firestoreContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to run container: %v", err)
		return
	}

	projectID := firestoreContainer.ProjectID()

	conn, err := grpc.NewClient(firestoreContainer.URI(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(emulatorCreds{}))
	if err != nil {
		log.Printf("failed to dial: %v", err)
		return
	}

	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := datastore.NewClient(ctx, projectID, options...)
	if err != nil {
		log.Printf("failed to create client: %v", err)
		return
	}
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
	if err != nil {
		log.Printf("failed to create entity: %v", err)
		return
	}

	saved := Person{}
	err = client.Get(ctx, userKey, &saved)
	if err != nil {
		log.Printf("failed to get entity: %v", err)
		return
	}

	fmt.Println(saved.Firstname, saved.Lastname)

	// Output:
	// Ada Lovelace
}
