package mongodb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
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
		{
			name: "With Replica set and mongo:7",
			img:  "mongo:7",
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
			testcontainers.CleanupContainer(t, mongodbContainer)
			require.NoError(tt, err)

			endpoint, err := mongodbContainer.ConnectionString(ctx)
			require.NoError(tt, err)

			// Force direct connection to the container to avoid the replica set
			// connection string that is returned by the container itself when
			// using the replica set option.
			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint).SetDirect(true))
			require.NoError(tt, err)

			err = mongoClient.Ping(ctx, nil)
			require.NoError(tt, err)
			require.Equal(t, "test", mongoClient.Database("test").Name())

			_, err = mongoClient.Database("testcontainer").Collection("test").InsertOne(context.Background(), bson.M{})
			require.NoError(tt, err)
		})
	}
}
