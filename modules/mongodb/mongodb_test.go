package mongodb_test

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func getLocalNonLoopbackIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		// Skip down or loopback interfaces.
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue // try next interface
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Check if it's a valid IPv4 and not loopback.
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not IPv4
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("no non-loopback IP address found")
}
func TestMongoDB(t *testing.T) {
	host, err := getLocalNonLoopbackIP()
	if err != nil {
		host = "host.docker.internal"
	}
	os.Setenv("TESTCONTAINERS_HOST_OVERRIDE", host)
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

func TestMongoDBChangeStream(t *testing.T) {
	host, err := getLocalNonLoopbackIP()
	if err != nil {
		host = "host.docker.internal"
	}
	os.Setenv("TESTCONTAINERS_HOST_OVERRIDE", host)

	ctx := context.Background()

	// Start MongoDB with replica set (required for change streams)
	mongodbContainer, err := mongodb.Run(ctx, "mongo:7",
		mongodb.WithReplicaSet("rs0"),
	)
	require.NoError(t, err)
	testcontainers.CleanupContainer(t, mongodbContainer)

	endpoint, err := mongodbContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	require.NoError(t, err)
	defer mongoClient.Disconnect(ctx)

	// Create a collection
	coll := mongoClient.Database("test").Collection("changes")

	// Start change stream
	stream, err := coll.Watch(ctx, mongo.Pipeline{})
	require.NoError(t, err)
	defer stream.Close(ctx)

	// Insert a document
	doc := bson.M{"message": "hello change streams"}
	_, err = coll.InsertOne(ctx, doc)
	require.NoError(t, err)

	// Wait for the change event
	require.True(t, stream.Next(ctx))

	var changeEvent bson.M
	err = stream.Decode(&changeEvent)
	require.NoError(t, err)

	// Verify the change event
	operationType, ok := changeEvent["operationType"].(string)
	require.True(t, ok)
	require.Equal(t, "insert", operationType)

	fullDocument, ok := changeEvent["fullDocument"].(bson.M)
	require.True(t, ok)
	require.Equal(t, "hello change streams", fullDocument["message"])
}
