package mongodb_test

import (
	"context"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestMongoDB(t *testing.T) {
	type tests struct {
		name  string
		image string
	}
	testCases := []tests{
		{
			name:  "From Docker Hub",
			image: "mongo:6",
		},
		{
			name:  "Community Server",
			image: "mongodb/mongodb-community-server:7.0.2-ubi8",
		},
		{
			name:  "Enterprise Server",
			image: "mongodb/mongodb-enterprise-server:7.0.0-ubi8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.image, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			mongodbContainer, err := mongodb.RunContainer(ctx, testcontainers.WithImage(tc.image))
			if err != nil {
				t.Fatalf("failed to start container: %s", err)
			}

			defer func() {
				if err := mongodbContainer.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			}()

			endpoint, err := mongodbContainer.ConnectionString(ctx)
			if err != nil {
				t.Fatalf("failed to get connection string: %s", err)
			}

			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
			if err != nil {
				t.Fatalf("failed to connect to MongoDB: %s", err)
			}

			err = mongoClient.Ping(ctx, nil)
			if err != nil {
				log.Fatalf("failed to ping MongoDB: %s", err)
			}

			if mongoClient.Database("test").Name() != "test" {
				t.Fatalf("failed to connect to the correct database")
			}
		})
	}
}
