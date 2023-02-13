package mongodb

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDB(t *testing.T) {
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

	// perform assertions

	endpoint, err := container.Endpoint(ctx, "mongodb")
	if err != nil {
		t.Error(fmt.Errorf("failed to get endpoint: %w", err))
	}

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(endpoint))
	if err != nil {
		t.Fatal(fmt.Errorf("error creating mongo client: %w", err))
	}

	err = mongoClient.Connect(ctx)
	if err != nil {
		t.Fatal(fmt.Errorf("error connecting to mongo: %w", err))
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		t.Fatal(fmt.Errorf("error pinging mongo: %w", err))
	}
}
