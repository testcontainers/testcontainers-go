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
		name string
		opts []testcontainers.ContainerCustomizer
	}
	testCases := []tests{
		{
			name: "From Docker Hub",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("mongo:6"),
			},
		},
		{
			name: "Community Server",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("mongodb/mongodb-community-server:7.0.2-ubi8"),
			},
		},
		{
			name: "Enterprise Server",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("mongodb/mongodb-enterprise-server:7.0.0-ubi8"),
			},
		},
		{
			name: "With Replica set and mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("mongo:4"),
				mongodb.WithReplicaSet(),
			},
		},
		{
			name: "With Replica set and mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				testcontainers.WithImage("mongo:6"),
				mongodb.WithReplicaSet(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.Background()

			mongodbContainer, err := mongodb.RunContainer(ctx, tc.opts...)
			if err != nil {
				tt.Fatalf("failed to start container: %s", err)
			}

			defer func() {
				if err := mongodbContainer.Terminate(ctx); err != nil {
					tt.Fatalf("failed to terminate container: %s", err)
				}
			}()

			endpoint, err := mongodbContainer.ConnectionString(ctx)
			if err != nil {
				tt.Fatalf("failed to get connection string: %s", err)
			}

			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
			if err != nil {
				tt.Fatalf("failed to connect to MongoDB: %s", err)
			}

			err = mongoClient.Ping(ctx, nil)
			if err != nil {
				log.Fatalf("failed to ping MongoDB: %s", err)
			}

			if mongoClient.Database("test").Name() != "test" {
				tt.Fatalf("failed to connect to the correct database")
			}
		})
	}
}
