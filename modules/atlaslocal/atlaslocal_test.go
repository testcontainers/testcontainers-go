package atlaslocal_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

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

func TestWithAuth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create username and password files.
	usernameFilepath := filepath.Join(tmpDir, "username.txt")

	err := os.WriteFile(usernameFilepath, []byte("file_testuser"), 0755)
	require.NoError(t, err)

	_, err = os.Stat(usernameFilepath)
	require.NoError(t, err, "Username file should exist")

	// Create the password file.
	passwordFilepath := filepath.Join(tmpDir, "password.txt")

	err = os.WriteFile(passwordFilepath, []byte("file_testpass"), 0755)
	require.NoError(t, err)

	_, err = os.Stat(passwordFilepath)
	require.NoError(t, err, "Password file should exist")

	cases := []struct {
		name    string
		auth    atlaslocal.AuthConfig
		creds   *options.Credential
		wantErr error
	}{
		{
			name:    "without auth",
			auth:    atlaslocal.AuthConfig{},
			creds:   nil,
			wantErr: nil,
		},
		{
			name:    "with auth",
			auth:    atlaslocal.AuthConfig{Username: "testuser", Password: "testpass"},
			creds:   &options.Credential{Username: "testuser", Password: "testpass"},
			wantErr: nil,
		},
		{
			name: "with auth files",
			auth: atlaslocal.AuthConfig{
				UsernameFile: filepath.Join(tmpDir, "username.txt"),
				PasswordFile: filepath.Join(tmpDir, "password.txt"),
			},
			creds:   &options.Credential{Username: "file_testuser", Password: "file_testpass"},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the container with the specified authentication configuration.
			opts := []testcontainers.ContainerCustomizer{atlaslocal.WithAuth(tc.auth)}

			ctr, err := atlaslocal.Run(context.Background(), latestImage, opts...)
			require.NoError(t, err)

			testcontainers.CleanupContainer(t, ctr)

			defer func() {
				err := ctr.Terminate(context.Background())
				require.NoError(t, err)
			}()

			// Verify the environment variables are set correctly.
			assertEnvVar(t, ctr, "MONGODB_INITDB_ROOT_USERNAME", tc.auth.Username)
			assertEnvVar(t, ctr, "MONGODB_INITDB_ROOT_PASSWORD", tc.auth.Password)
			assertEnvVar(t, ctr, "MONGODB_INITDB_ROOT_USERNAME_FILE", tc.auth.UsernameFile)
			assertEnvVar(t, ctr, "MONGODB_INITDB_ROOT_PASSWORD_FILE", tc.auth.PasswordFile)

			// Connect to the MongoDB Atlas Local instance using the provided
			// credentials.
			clientOpts := options.Client().ApplyURI(createConnectionURI(t, context.Background(), ctr).String())
			if tc.creds != nil {
				clientOpts.SetAuth(*tc.creds)
			}

			client, err := mongo.Connect(clientOpts)
			require.NoError(t, err)

			defer func() {
				err := client.Disconnect(context.Background())
				require.NoError(t, err)
			}()

			// Execute an insert operation to verify the connection and
			// authentication.
			coll := client.Database("test").Collection("foo")

			_, err = coll.InsertOne(context.Background(), bson.D{{Key: "test", Value: "value"}})
			require.NoError(t, err, "Failed to insert document with authentication")
		})
	}
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

func TestWithInitDatabase(t *testing.T) {
	initScripts := map[string]string{
		"01-seed.js": `db.foo.insertOne({ _id: 1, seeded: true });`,
	}

	tmpDir := createInitScripts(t, initScripts)
	opts := []testcontainers.ContainerCustomizer{
		atlaslocal.WithInitDatabase("mydb"),
		atlaslocal.WithInitScripts(tmpDir),
	}

	ctr, err := atlaslocal.Run(context.Background(), latestImage, opts...)
	require.NoError(t, err)

	testcontainers.CleanupContainer(t, ctr)

	defer func() {
		err := ctr.Terminate(context.Background())
		require.NoError(t, err)
	}()

	assertInitScriptsExist(t, ctr, tmpDir, initScripts)
	assertEnvVar(t, ctr, "MONGODB_INITDB_DATABASE", "mydb")

	client := newMongoClient(t, context.Background(), ctr)
	defer func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err)
	}()

	coll := client.Database("mydb").Collection("foo")

	res := coll.FindOne(context.Background(), bson.D{{"_id", int32(1)}})
	require.NoError(t, res.Err())

	var doc bson.D
	require.NoError(t, res.Decode(&doc), "Failed to decode seeded document")
	assert.Equal(t, bson.D{{"_id", int32(1)}, {"seeded", true}}, doc, "Seeded document does not match expected values")
}

func TestWithInitScripts(t *testing.T) {
	cases := []struct {
		name        string
		initScripts map[string]string // filename -> content
		want        []bson.D
	}{
		{
			name:        "no scripts",
			initScripts: map[string]string{},
			want:        []bson.D{},
		},
		{
			name: "single shell script",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
			},
			want: []bson.D{{{"_id", int32(1)}, {"seeded", true}}},
		},
		{
			name: "single js script",
			initScripts: map[string]string{
				"01-seed.js": `db.foo.insertOne({ _id: 1, seeded: true });`,
			},
			want: []bson.D{{{"_id", int32(1)}, {"seeded", true}}},
		},
		{
			name: "mixed shell and js scripts",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
				"02-seed.js": `db.foo.insertOne({ _id: 2, seeded: true });`,
			},
			want: []bson.D{
				{{"_id", int32(1)}, {"seeded", true}},
				{{"_id", int32(2)}, {"seeded", true}},
			},
		},
		{
			name: "mixed ordered shell",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
				"02-seed.sh": `mongosh --eval 'db.foo.deleteOne({ _id: 1, seeded: true })'`,
			},
			want: []bson.D{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := createInitScripts(t, tc.initScripts)

			// Start container with the init scripts mounted.
			opts := []testcontainers.ContainerCustomizer{
				atlaslocal.WithInitScripts(tmpDir),
			}

			ctr, err := atlaslocal.Run(context.Background(), latestImage, opts...)
			require.NoError(t, err)

			testcontainers.CleanupContainer(t, ctr)

			defer func() {
				err := ctr.Terminate(context.Background())
				require.NoError(t, err)
			}()

			assertInitScriptsExist(t, ctr, tmpDir, tc.initScripts)

			// Connect to the server.
			client := newMongoClient(t, context.Background(), ctr)
			defer func() {
				err := client.Disconnect(context.Background())
				require.NoError(t, err)
			}()

			// Fetch the seeded data.
			coll := client.Database("test").Collection("foo")

			cur, err := coll.Find(context.Background(), bson.D{})
			require.NoError(t, err)

			var results []bson.D
			require.NoError(t, cur.All(context.Background(), &results))

			assert.ElementsMatch(t, results, tc.want, "Seeded documents do not match expected values")
		})
	}
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

// dumpLogs will dump the logs of the MongoDB Atlas Local container to the
// integration test output.
func dumpLogs(t *testing.T, ctx context.Context, ctr testcontainers.Container) {
	t.Helper()

	r, err := ctr.Logs(ctx)
	require.NoError(t, err)

	bytes, err := io.ReadAll(r)
	t.Logf("MongoDB Atlas Local logs:\n%s", string(bytes))
}

// executeAggregation connects to the MongoDB Atlas Local instance, creates a
// collection with a search index, inserts a document, and performs an
// aggregation using the search index.
func executeAggregation(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Connect to a MongoDB Atlas Local instance and create a collection with a
	// search index.
	client, err := mongo.Connect(options.Client().ApplyURI(createConnectionURI(t, context.Background(), ctr).String()))
	require.NoError(t, err)

	//// Await the server to be ready.
	//require.Eventually(t, func() bool {
	//	pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	//	defer cancel()

	//	err := client.Ping(pingCtx, nil)
	//	return err == nil
	//}, 60*time.Second, 5*time.Second)

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

// TODO: remove this?
func newMongoClient(t *testing.T, ctx context.Context, ctr testcontainers.Container) *mongo.Client {
	t.Helper()
	// Connect to the server.
	uri := createConnectionURI(t, context.Background(), ctr)

	client, err := mongo.Connect(options.Client().ApplyURI(uri.String()))
	require.NoError(t, err)

	return client
}

func createInitScripts(t *testing.T, scripts map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()

	for filename, content := range scripts {
		scriptPath := filepath.Join(tmpDir, filename)
		require.NoError(t, os.WriteFile(scriptPath, []byte(content), 0755))

		// Sanity check to verify that the script content is as expected.
		got, err := os.ReadFile(scriptPath)
		require.NoError(t, err, "Failed to read init script %s", filename)
		assert.Equal(t, string(got), content, "Content of init script %s does not match", filename)
	}

	return tmpDir
}

func assertInitScriptsExist(t *testing.T, ctr testcontainers.Container, tmpDir string, expectedScripts map[string]string) {
	t.Helper()

	// Sanity check on the container's binds.
	insp, err := ctr.Inspect(context.Background())
	require.NoError(t, err)
	require.Len(t, insp.HostConfig.Binds, 1, "Expected exactly one bind mount for init scripts")

	want := fmt.Sprintf("%s:/docker-entrypoint-initdb.d:ro", tmpDir)
	assert.Equal(t, want, insp.HostConfig.Binds[0])

	// Sanity check to verify that all scripts are present.
	for filename := range expectedScripts {
		cmd := []string{"sh", "-c", fmt.Sprintf("cd docker-entrypoint-initdb.d && ls -l")}

		exitCode, reader, err := ctr.Exec(context.Background(), cmd)
		require.NoError(t, err)
		require.Zero(t, exitCode, "Expected exit code 0 for command: %v", cmd)

		content, _ := io.ReadAll(reader)
		assert.Contains(t, string(content), filename)
	}
}
