package mongodb_test

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestMongoDB(t *testing.T) {
	type tests struct {
		name string
		img  string
		opts []testcontainers.ContainerCustomizer
	}
	testCases := []tests{
		{
			name: "From Docker Hub",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "Community Server",
			img:  "mongodb/mongodb-community-server:7.0.2-ubi8",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "Enterprise Server",
			img:  "mongodb/mongodb-enterprise-server:7.0.0-ubi8",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "With Replica set and mongo:4",
			img:  "mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "With Replica set and mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.Background()

			mongodbContainer, err := mongodb.Run(ctx, tc.img, tc.opts...)
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

			// Force direct connection to the container to avoid the replica set
			// connection string that is returned by the container itself when
			// using the replica set option.
			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint+"/?connect=direct"))
			if err != nil {
				tt.Fatalf("failed to connect to MongoDB: %s", err)
			}

			err = mongoClient.Ping(ctx, nil)
			if err != nil {
				tt.Fatalf("failed to ping MongoDB: %s", err)
			}

			if mongoClient.Database("test").Name() != "test" {
				tt.Fatalf("failed to connect to the correct database")
			}
		})
	}
}
