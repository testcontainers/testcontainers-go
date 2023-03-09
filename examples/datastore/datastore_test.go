package datastore

import (
	"cloud.google.com/go/datastore"
	"context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
)

type Task struct {
	Description string
}

func TestDatastore(t *testing.T) {
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

	options := []option.ClientOption{
		option.WithEndpoint(container.URI),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	dsClient, err := datastore.NewClient(ctx, "test-project", options...)
	if err != nil {
		t.Fatal(err)
	}
	defer dsClient.Close()

	k := datastore.NameKey("Task", "sample", nil)
	data := Task{
		Description: "my description",
	}
	_, err = dsClient.Put(ctx, k, &data)
	if err != nil {
		t.Fatal(err)
	}

	saved := Task{}
	err = dsClient.Get(ctx, k, &saved)
	if err != nil {
		t.Fatal(err)
	}

	// perform assertions
	if saved != data {
		t.Fatalf("Expected value %s. Got %s.", data, saved)
	}
}
