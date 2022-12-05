package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"testing"
)

type Person struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

func TestFirestore(t *testing.T) {
	ctx := context.Background()

	container, err := setupFirestore(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	t.Setenv("FIRESTORE_EMULATOR_HOST", container.URI)
	client, err := firestore.NewClient(ctx, "test-project")
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
