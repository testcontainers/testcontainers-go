package mongodb_test

import (
	"context"
	"fmt"
	"net/url"
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
			name: "with-replica/mongo:4",
			img:  "mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-replica/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-replica/mongo:7",
			img:  "mongo:7",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
			},
		},
		{
			name: "with-auth/replica/mongo:7",
			img:  "mongo:7",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/mongo:6",
			img:  "mongo:6",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongodb-enterprise-server:7.0.0-ubi8",
			img:  "mongodb/mongodb-enterprise-server:7.0.0-ubi8",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongodb-community-server:7.0.2-ubi8",
			img:  "mongodb/mongodb-community-server:7.0.2-ubi8",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
			},
		},
		{
			name: "with-auth/replica/mongo:4",
			img:  "mongo:4",
			opts: []testcontainers.ContainerCustomizer{
				mongodb.WithReplicaSet("rs"),
				mongodb.WithUsername("tester"),
				mongodb.WithPassword("testerpass"),
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

			// Force direct connection to the container.
			mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint).SetDirect(true))
			require.NoError(tt, err)

			err = mongoClient.Ping(ctx, nil)
			require.NoError(tt, err)
			require.Equal(tt, "test", mongoClient.Database("test").Name())

			// Basic insert test.
			_, err = mongoClient.Database("testcontainer").Collection("test").InsertOne(ctx, bson.M{})
			require.NoError(tt, err)

			// If the container is configured with a replica set, run the change stream test.
			if hasReplica, _ := hasReplicaSet(endpoint); hasReplica {
				coll := mongoClient.Database("test").Collection("changes")
				stream, err := coll.Watch(ctx, mongo.Pipeline{})
				require.NoError(tt, err)
				defer stream.Close(ctx)

				doc := bson.M{"message": "hello change streams"}
				_, err = coll.InsertOne(ctx, doc)
				require.NoError(tt, err)

				require.True(tt, stream.Next(ctx))
				var changeEvent bson.M
				err = stream.Decode(&changeEvent)
				require.NoError(tt, err)

				opType, ok := changeEvent["operationType"].(string)
				require.True(tt, ok, "Expected operationType field")
				require.Equal(tt, "insert", opType, "Expected operationType to be 'insert'")

				fullDoc, ok := changeEvent["fullDocument"].(bson.M)
				require.True(tt, ok, "Expected fullDocument field")
				require.Equal(tt, "hello change streams", fullDoc["message"])
			}
		})
	}
}

// hasReplicaSet checks if the connection string includes a replicaSet query parameter.
func hasReplicaSet(connStr string) (bool, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return false, fmt.Errorf("parse connection string: %w", err)
	}
	q := u.Query()
	return q.Get("replicaSet") != "", nil
}
