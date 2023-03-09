package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

type Person struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type emulatorCreds struct {
}

func (ec emulatorCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer owner"}, nil
}
func (ec emulatorCreds) RequireTransportSecurity() bool {
	return false
}

func TestFirestore(t *testing.T) {
	ctx := context.Background()

	container, err := startContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	conn, err := grpc.Dial(container.URI, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(emulatorCreds{}))
	if err != nil {
		t.Fatal(err)
	}
	options := []option.ClientOption{option.WithGRPCConn(conn)}
	client, err := firestore.NewClient(ctx, "test-project", options...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	users := client.Collection("users")
	docRef := users.Doc("alovelace")

	data := Person{
		Firstname: "Ada",
		Lastname:  "Lovelace",
	}
	_, err = docRef.Create(ctx, data)
	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	docsnap, err := docRef.Get(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var saved Person
	if err := docsnap.DataTo(&saved); err != nil {
		t.Fatal(err)
	}

	if saved != data {
		t.Fatalf("Expected value %s. Got %s.", data, saved)
	}
}
