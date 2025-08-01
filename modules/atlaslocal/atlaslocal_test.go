package atlaslocal_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gotest.tools/v3/assert"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/atlaslocal"
)

const latestImage = "mongodb/mongodb-atlas-local:latest"

func TestMongoDBAtlasLocal(t *testing.T) {
	ctx := context.Background()

	ctr, err := atlaslocal.Run(ctx, latestImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	client, err := mongo.Connect(options.Client().ApplyURI(createConnectionURI(t, ctx, ctr).String()))
	require.NoError(t, err)

	err = client.Ping(ctx, nil)
	require.NoError(t, err)
}

func TestMongoDBAtlasLocal_WithDisableTelemetry(t *testing.T) {
	t.Run("with", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithDisableTelemetry())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "DO_NOT_TRACK", "1")
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "DO_NOT_TRACK", "")
	})
}

func TestMongoDBAtlasLocal_WithMongotLogFile(t *testing.T) {
	const mongotLogFile = "/tmp/mongot.log"

	t.Run("with", func(t *testing.T) {
		// This test is to ensure that the DO_NOT_TRACK environment variable is set correctly
		// when the WithDisableTelemetry option is used.
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithMongotLogFile(mongotLogFile))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "MONGOT_LOG_FILE", mongotLogFile)

		executeAggregation(t, ctr)

		assertMongotLog(t, ctr, mongotLogFile, false)
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "MONGOT_LOG_FILE", "")
		executeAggregation(t, ctr)

		assertMongotLog(t, ctr, mongotLogFile, true)
	})
}

func TestMongoDBAtlasLocal_WithRunnerLogFile(t *testing.T) {
	const runnerLogFile = "/tmp/runner.log"

	t.Run("with", func(t *testing.T) {
		// This test is to ensure that the DO_NOT_TRACK environment variable is set correctly
		// when the WithDisableTelemetry option is used.
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithRunnerLogFile(runnerLogFile))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "RUNNER_LOG_FILE", runnerLogFile)
		assertMongotLog(t, ctr, runnerLogFile, false)
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		defer func() {
			err := ctr.Terminate(context.Background())
			require.NoError(t, err)
		}()

		assertEnvVar(t, ctr, "RUNNER_LOG_FILE", "")
		executeAggregation(t, ctr)

		assertMongotLog(t, ctr, runnerLogFile, true)
	})
}

// Test Helper Functions

func assertEnvVar(t *testing.T, ctr testcontainers.Container, envVarName, expected string) {
	t.Helper()

	exitCode, reader, err := ctr.Exec(context.Background(), []string{"sh", "-c", fmt.Sprintf("echo $%s", envVarName)})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	outBytes, err := io.ReadAll(reader)
	require.NoError(t, err)

	if len(outBytes) < 8 {
		t.Fatalf("Exec output too short: less than 8 bytes!")
	}

	out := strings.TrimSpace(string(outBytes[8:]))
	assert.Equal(t, expected, out, "DO_NOT_TRACK env var value mismatch")
}

func assertMongotLog(t *testing.T, ctr testcontainers.Container, mongotLogFile string, empty bool) {
	t.Helper()

	// Pull the log file and assert non-empty.
	reader, _ := ctr.CopyFileFromContainer(context.Background(), mongotLogFile)

	var data []byte
	if reader != nil {
		data, _ = io.ReadAll(reader)
	}

	if empty {
		require.Empty(t, data, "mongot log file should be empty")
	} else {
		require.NotEmpty(t, data, "mongot log file should not be empty")
	}
}

func createConnectionURI(t *testing.T, ctx context.Context, ctr testcontainers.Container) *url.URL {
	t.Helper()

	// perform assertions
	host, err := ctr.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := ctr.MappedPort(ctx, "27017")
	require.NoError(t, err)

	uri := &url.URL{
		Scheme:   "mongodb",
		Host:     net.JoinHostPort(host, mappedPort.Port()),
		Path:     "/",
		RawQuery: "directConnection=true",
	}

	return uri
}

// createSeachIndex creates a search index with the given name on the provided
// collection and waits for it to be acknowledged server-side.
func createSeachIndex(t *testing.T, ctx context.Context, coll *mongo.Collection, indexName string) {
	t.Helper()

	// Create the default definition for search index
	definition := bson.D{{Key: "mappings", Value: bson.D{{Key: "dynamic", Value: true}}}}
	indexModel := mongo.SearchIndexModel{
		Definition: definition,
		Options:    options.SearchIndexes().SetName(indexName),
	}

	_, err := coll.SearchIndexes().CreateOne(ctx, indexModel)
	require.NoError(t, err)
}

// executeAggregation connects to the MongoDB Atlas Local instance, creates a
// collection with a search index, inserts a document, and performs an
// aggregation using the search index.
func executeAggregation(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// TODO: Figure out how to remove this.
	// Wait for the container to start and log file to be created
	time.Sleep(10 * time.Second)

	// Connect to a MongoDB Atlas Local instance and create a collection with a
	// search index.
	client, err := mongo.Connect(options.Client().ApplyURI(createConnectionURI(t, context.Background(), ctr).String()))
	require.NoError(t, err)

	err = client.Database("test").CreateCollection(context.Background(), "search")
	require.NoError(t, err)

	coll := client.Database("test").Collection("search")

	siCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a search index on the collection.
	createSeachIndex(t, siCtx, coll, "test_search_index")

	// Insert a document into the collection and aggregate it using the search
	// index which should log the operation to the mongot log file.
	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "txt", Value: "hello"}})
	require.NoError(t, err)

	pipeline := mongo.Pipeline{{
		{"$search", bson.D{
			{"text", bson.D{{"query", "hello"}, {"path", "txt"}}},
		}},
	}}

	cur, err := coll.Aggregate(context.Background(), pipeline)
	require.NoError(t, err)

	err = cur.Close(context.Background())
	require.NoError(t, err)
}
