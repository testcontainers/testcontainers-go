package mongodb

import (
	"context"
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongodb(t *testing.T) {
	ctx := context.Background()

	container, err := setupMongodb(ctx)
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

	host, _ := container.Host(ctx)

	p, _ := container.MappedPort(ctx, "27017/tcp")
	port := p.Int()

	connectionString := fmt.Sprintf("mongodb://%s:%d", host, port)

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
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
