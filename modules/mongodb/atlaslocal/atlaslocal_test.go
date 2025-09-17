package atlaslocal_test

import (
	"context"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/x/mongo/driver/connstring"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb/atlaslocal"
)

const latestImage = "mongodb/mongodb-atlas-local:latest"

func TestMongoDBAtlasLocal(t *testing.T) {
	ctx := context.Background()

	ctr, err := atlaslocal.Run(ctx, latestImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	client, td := newMongoClient(t, ctx, ctr)
	defer td()

	err = client.Ping(ctx, nil)
	require.NoError(t, err)
}

func TestSCRAMAuth(t *testing.T) {
	tmpDir, usernameFilepath, passwordFilepath := newAuthFiles(t)

	cases := []struct {
		name         string
		username     string
		password     string
		usernameFile string
		passwordFile string
		wantRunErr   string
	}{
		{
			name:         "without auth",
			username:     "",
			password:     "",
			usernameFile: "",
			passwordFile: "",
			wantRunErr:   "",
		},
		{
			name:         "with auth",
			username:     "testuser",
			password:     "testpass",
			usernameFile: "",
			passwordFile: "",
			wantRunErr:   "",
		},
		{
			name:         "with auth files",
			username:     "",
			password:     "",
			usernameFile: usernameFilepath,
			passwordFile: passwordFilepath,
			wantRunErr:   "",
		},
		{
			name:         "with inline and files",
			username:     "testuser",
			password:     "testpass",
			usernameFile: usernameFilepath,
			passwordFile: passwordFilepath,
			wantRunErr:   "you cannot specify both inline credentials and files for credentials",
		},
		{
			name:         "username without password",
			username:     "testuser",
			password:     "",
			usernameFile: "",
			passwordFile: "",
			wantRunErr:   "if you specify username or password, you must provide both of them",
		},
		{
			name:         "password without username",
			username:     "",
			password:     "testpass",
			usernameFile: "",
			passwordFile: "",
			wantRunErr:   "if you specify username or password, you must provide both of them",
		},
		{
			name:         "username file without password file",
			username:     "",
			password:     "",
			usernameFile: usernameFilepath,
			passwordFile: "",
			wantRunErr:   "if you specify username file or password file, you must provide both of them",
		},
		{
			name:         "password file without username file",
			username:     "",
			password:     "",
			usernameFile: "",
			passwordFile: passwordFilepath,
			wantRunErr:   "if you specify username file or password file, you must provide both of them",
		},
		{
			name:         "username file invalid mount path",
			username:     "",
			password:     "",
			usernameFile: "nonexistent_username.txt",
			passwordFile: passwordFilepath,
			wantRunErr:   "mount path must be absolute",
		},
		{
			name:         "password file invalid mount path",
			username:     "",
			password:     "",
			usernameFile: usernameFilepath,
			passwordFile: "nonexistent_password.txt",
			wantRunErr:   "mount path must be absolute",
		},
		{
			name:         "username file is absolute but does not exist",
			username:     "",
			password:     "",
			usernameFile: "/nonexistent/username.txt",
			passwordFile: passwordFilepath,
			wantRunErr:   "does not exist or is not accessible",
		},
		{
			name:         "password file is absolute but does not exist",
			username:     "",
			password:     "",
			usernameFile: usernameFilepath,
			passwordFile: "/nonexistent/password.txt",
			wantRunErr:   "does not exist or is not accessible",
		},
		{
			name:         "username file is a directory",
			username:     "",
			password:     "",
			usernameFile: tmpDir,
			passwordFile: passwordFilepath,
			wantRunErr:   "must be a file",
		},
		{
			name:         "password file is a directory",
			username:     "",
			password:     "",
			usernameFile: usernameFilepath,
			passwordFile: tmpDir,
			wantRunErr:   "must be a file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Construct the custom options for the MongoDB Atlas Local container.
			opts := []testcontainers.ContainerCustomizer{}

			if tc.username != "" {
				opts = append(opts, atlaslocal.WithUsername(tc.username))
			}

			if tc.password != "" {
				opts = append(opts, atlaslocal.WithPassword(tc.password))
			}

			if tc.usernameFile != "" {
				opts = append(opts, atlaslocal.WithUsernameFile(tc.usernameFile))
			}

			if tc.passwordFile != "" {
				opts = append(opts, atlaslocal.WithPasswordFile(tc.passwordFile))
			}

			// Randomize the order of the options
			rand.Shuffle(len(opts), func(i, j int) {
				opts[i], opts[j] = opts[j], opts[i]
			})

			// Create the MongoDB Atlas Local container with the specified options.
			ctr, err := atlaslocal.Run(context.Background(), latestImage, opts...)
			testcontainers.CleanupContainer(t, ctr)

			if tc.wantRunErr != "" {
				require.ErrorContains(t, err, tc.wantRunErr)

				return
			}

			require.NoError(t, err)

			// Verify the environment variables are set correctly.
			requireEnvVar(t, ctr, "MONGODB_INITDB_ROOT_USERNAME", tc.username)
			requireEnvVar(t, ctr, "MONGODB_INITDB_ROOT_PASSWORD", tc.password)

			if tc.usernameFile != "" {
				requireEnvVar(t, ctr, "MONGODB_INITDB_ROOT_USERNAME_FILE", "/run/secrets/mongo-root-username")
			}

			if tc.passwordFile != "" {
				requireEnvVar(t, ctr, "MONGODB_INITDB_ROOT_PASSWORD_FILE", "/run/secrets/mongo-root-password")
			}

			client, td := newMongoClient(t, context.Background(), ctr)
			defer td()

			// Execute an insert operation to verify the connection and
			// authentication.
			coll := client.Database("test").Collection("foo")

			_, err = coll.InsertOne(context.Background(), bson.D{{Key: "test", Value: "value"}})
			require.NoError(t, err, "Failed to insert document with authentication")
		})
	}
}

func TestWithNoTelemetry(t *testing.T) {
	t.Run("with", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithNoTelemetry())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "DO_NOT_TRACK", "1")
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "DO_NOT_TRACK", "")
	})
}

func TestWithMongotLogFile(t *testing.T) {
	t.Run("with", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithMongotLogFile())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "MONGOT_LOG_FILE", "/tmp/mongot.log")

		executeAggregation(t, ctr)
		requireMongotLogs(t, ctr)
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "MONGOT_LOG_FILE", "")

		executeAggregation(t, ctr)
		requireNoMongotLogs(t, ctr)
	})

	t.Run("to stdout", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage,
			atlaslocal.WithMongotLogToStdout())
		testcontainers.CleanupContainer(t, ctr)

		require.NoError(t, err)

		requireEnvVar(t, ctr, "MONGOT_LOG_FILE", "/dev/stdout")

		executeAggregation(t, ctr)

		requireMongotLogs(t, ctr)
		requireContainerLogsNotEmpty(t, ctr)
	})

	t.Run("to stderr", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage,
			atlaslocal.WithMongotLogToStderr())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "MONGOT_LOG_FILE", "/dev/stderr")

		executeAggregation(t, ctr)

		requireMongotLogs(t, ctr)
		requireContainerLogsNotEmpty(t, ctr)
	})
}

func TestWithRunnerLogFile(t *testing.T) {
	const runnerLogFile = "/tmp/runner.log"

	t.Run("with", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage, atlaslocal.WithRunnerLogFile())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "RUNNER_LOG_FILE", runnerLogFile)
		requireRunnerLogs(t, ctr)
	})

	t.Run("without", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage)
		testcontainers.CleanupContainer(t, ctr)

		require.NoError(t, err)

		requireEnvVar(t, ctr, "RUNNER_LOG_FILE", "")
		requireNoRunnerLogs(t, ctr)
	})

	t.Run("to stdout", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage,
			atlaslocal.WithRunnerLogToStdout())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "RUNNER_LOG_FILE", "/dev/stdout")

		executeAggregation(t, ctr)

		requireRunnerLogs(t, ctr)
		requireContainerLogsNotEmpty(t, ctr)
	})

	t.Run("to stderr", func(t *testing.T) {
		ctr, err := atlaslocal.Run(context.Background(), latestImage,
			atlaslocal.WithRunnerLogToStderr())
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		requireEnvVar(t, ctr, "RUNNER_LOG_FILE", "/dev/stderr")

		executeAggregation(t, ctr)

		requireRunnerLogs(t, ctr)
		requireContainerLogsNotEmpty(t, ctr)
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
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	requireInitScriptsExist(t, ctr, initScripts)
	requireEnvVar(t, ctr, "MONGODB_INITDB_DATABASE", "mydb")

	client, td := newMongoClient(t, context.Background(), ctr)
	defer td()

	coll := client.Database("mydb").Collection("foo")

	seed := bson.D{{Key: "_id", Value: int32(1)}, {Key: "seeded", Value: true}}

	res := coll.FindOne(context.Background(), seed)
	require.NoError(t, res.Err())

	var doc bson.D
	require.NoError(t, res.Decode(&doc), "Failed to decode seeded document")
	require.Equal(t, seed, doc, "Seeded document does not match expected values")
}

func TestWithInitScripts(t *testing.T) {
	seed1 := bson.D{{Key: "_id", Value: int32(1)}, {Key: "seeded", Value: true}}
	seed2 := bson.D{{Key: "_id", Value: int32(2)}, {Key: "seeded", Value: true}}

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
			want: []bson.D{seed1},
		},
		{
			name: "single js script",
			initScripts: map[string]string{
				"01-seed.js": `db.foo.insertOne({ _id: 1, seeded: true });`,
			},
			want: []bson.D{seed1},
		},
		{
			name: "mixed shell and js scripts",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
				"02-seed.js": `db.foo.insertOne({ _id: 2, seeded: true });`,
			},
			want: []bson.D{seed1, seed2},
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
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			requireInitScriptsExist(t, ctr, tc.initScripts)

			// Connect to the server.
			client, td := newMongoClient(t, context.Background(), ctr)
			defer td()

			// Fetch the seeded data.
			coll := client.Database("test").Collection("foo")

			cur, err := coll.Find(context.Background(), bson.D{})
			require.NoError(t, err)

			var results []bson.D
			require.NoError(t, cur.All(context.Background(), &results))

			require.ElementsMatch(t, results, tc.want, "Seeded documents do not match expected values")
		})
	}
}

// Ensure that we can chain multiple scripts.
func TestWithInitScripts_MultipleScripts(t *testing.T) {
	scripts1 := map[string]string{
		"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
		"02-seed.js": `db.foo.insertOne({ _id: 2, seeded: true });`,
	}

	tmpDir1 := createInitScripts(t, scripts1)

	scripts2 := map[string]string{
		"03-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 3, seeded: true })'`,
	}

	tmpDir2 := createInitScripts(t, scripts2)

	// Start container with the init scripts mounted.
	opts := []testcontainers.ContainerCustomizer{
		atlaslocal.WithInitScripts(tmpDir1),
		atlaslocal.WithInitScripts(tmpDir2),
	}

	ctr, err := atlaslocal.Run(context.Background(), latestImage, opts...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	requireInitScriptsDoesNotExist(t, ctr, scripts1)
	requireInitScriptsExist(t, ctr, scripts2)
}

func TestConnectionString(t *testing.T) {
	_, usernameFilepath, passwordFilepath := newAuthFiles(t)

	testcases := []struct {
		name         string
		opts         []testcontainers.ContainerCustomizer
		wantUsername string
		wantPassword string
		wantDatabase string
	}{
		{
			name:         "default",
			opts:         []testcontainers.ContainerCustomizer{},
			wantUsername: "",
			wantPassword: "",
			wantDatabase: "",
		},
		{
			name: "with auth options",
			opts: []testcontainers.ContainerCustomizer{
				atlaslocal.WithUsername("testuser"),
				atlaslocal.WithPassword("testpass"),
				atlaslocal.WithInitDatabase("testdb"),
			},
			wantUsername: "testuser",
			wantPassword: "testpass",
			wantDatabase: "testdb",
		},
		{
			name: "with auth files",
			opts: []testcontainers.ContainerCustomizer{
				atlaslocal.WithUsernameFile(usernameFilepath),
				atlaslocal.WithPasswordFile(passwordFilepath),
				atlaslocal.WithInitDatabase("testdb"),
			},
			wantUsername: "file_testuser",
			wantPassword: "file_testpass",
			wantDatabase: "testdb",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctr, err := atlaslocal.Run(context.Background(), latestImage, tc.opts...)
			require.NoError(t, err)

			csRaw, err := ctr.ConnectionString(context.Background())
			require.NoError(t, err)

			connString, err := connstring.ParseAndValidate(csRaw)
			require.NoError(t, err, "Failed to parse connection string")

			require.Equal(t, "mongodb", connString.Scheme)
			require.Equal(t, "localhost", connString.Hosts[0][:9])
			require.NotEmpty(t, connString.Hosts[0][10:], "Port should be non-empty")
			require.Equal(t, tc.wantUsername, connString.Username)
			require.Equal(t, tc.wantPassword, connString.Password)
			require.Equal(t, tc.wantDatabase, connString.Database)
			require.True(t, connString.DirectConnection)
		})
	}
}

// Test Helper Functions

func requireEnvVar(t *testing.T, ctr testcontainers.Container, envVarName, expected string) {
	t.Helper()

	exitCode, reader, err := ctr.Exec(context.Background(), []string{"sh", "-c", "echo $" + envVarName})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	outBytes, err := io.ReadAll(reader)
	require.NoError(t, err)

	// testcontainers-go's Exec() returns a multiplexed stream in the same format
	// used by the Docker API. Each frame is prefixed with an 8-byte header.
	require.Greater(t, len(outBytes), 8, "Exec output too short to contain env var value")

	out := strings.TrimSpace(string(outBytes[8:]))
	require.Equal(t, expected, out, "DO_NOT_TRACK env var value mismatch")
}

func requireMongotLogs(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Pull the log file and require non-empty.
	reader, err := ctr.(*atlaslocal.Container).ReadMongotLogs(context.Background())
	require.NoError(t, err)

	buf := make([]byte, 1)
	_, _ = reader.Read(buf) // read at least one byte to ensure non-empty
}

func requireNoMongotLogs(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Pull the log file and require non-empty.
	reader, err := ctr.(*atlaslocal.Container).ReadMongotLogs(context.Background())
	require.ErrorIs(t, err, os.ErrNotExist)

	if reader != nil { // Failure case where reader is non-nil
		_ = reader.Close()
	}
}

func requireRunnerLogs(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Pull the log file and require non-empty.
	reader, err := ctr.(*atlaslocal.Container).ReadRunnerLogs(context.Background())
	require.NoError(t, err)

	defer reader.Close()

	buf := make([]byte, 1)
	_, _ = reader.Read(buf) // Read at least one byte to ensure non-empty
}

func requireNoRunnerLogs(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Pull the log file and require non-empty.
	reader, err := ctr.(*atlaslocal.Container).ReadRunnerLogs(context.Background())
	require.ErrorIs(t, err, os.ErrNotExist)

	if reader != nil { // Failure case where reader is non-nil
		_ = reader.Close()
	}
}

// createSearchIndex creates a search index with the given name on the provided
// collection and waits for it to be acknowledged server-side.
func createSearchIndex(t *testing.T, ctx context.Context, coll *mongo.Collection, indexName string) {
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

	client, td := newMongoClient(t, context.Background(), ctr)
	defer td()

	err := client.Database("test").CreateCollection(context.Background(), "search")
	require.NoError(t, err)

	coll := client.Database("test").Collection("search")

	siCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a search index on the collection.
	createSearchIndex(t, siCtx, coll, "test_search_index")

	// Insert a document into the collection and aggregate it using the search
	// index which should log the operation to the mongot log file.
	_, err = coll.InsertOne(context.Background(), bson.D{{Key: "txt", Value: "hello"}})
	require.NoError(t, err)

	pipeline := mongo.Pipeline{{
		{Key: "$search", Value: bson.D{
			{Key: "text", Value: bson.D{{Key: "query", Value: "hello"}, {Key: "path", Value: "txt"}}},
		}},
	}}

	cur, err := coll.Aggregate(context.Background(), pipeline)
	require.NoError(t, err)

	err = cur.Close(context.Background())
	require.NoError(t, err)
}

func newMongoClient(
	t *testing.T,
	ctx context.Context,
	ctr testcontainers.Container,
	opts ...*options.ClientOptions,
) (*mongo.Client, func()) {
	t.Helper()

	connString, err := ctr.(*atlaslocal.Container).ConnectionString(ctx)
	require.NoError(t, err)

	copts := []*options.ClientOptions{
		options.Client().ApplyURI(connString),
	}

	copts = append(copts, opts...)

	client, err := mongo.Connect(copts...)
	require.NoError(t, err)

	return client, func() {
		err := client.Disconnect(context.Background())
		require.NoError(t, err, "Failed to disconnect MongoDB client")
	}
}

func createInitScripts(t *testing.T, scripts map[string]string) string {
	t.Helper()

	tmpDir := t.TempDir()

	for filename, content := range scripts {
		scriptPath := filepath.Join(tmpDir, filename)
		require.NoError(t, os.WriteFile(scriptPath, []byte(content), 0o755))

		// Sanity check to verify that the script content is as expected.
		got, err := os.ReadFile(scriptPath)
		require.NoError(t, err, "Failed to read init script %s", filename)
		require.Equal(t, string(got), content, "Content of init script %s does not match", filename)
	}

	return tmpDir
}

func requireInitScriptsExist(t *testing.T, ctr testcontainers.Container, expectedScripts map[string]string) {
	t.Helper()

	const dstDir = "/docker-entrypoint-initdb.d"

	exit, r, err := ctr.Exec(context.Background(), []string{"sh", "-lc", "ls -l " + dstDir})
	require.NoError(t, err)

	// If the map is empty, the command returns exit code 2.
	if len(expectedScripts) == 0 {
		require.Equal(t, 2, exit)
	} else {
		require.Equal(t, 0, exit)
	}

	listingBytes, err := io.ReadAll(r)
	require.NoError(t, err)

	listing := string(listingBytes)

	for name, want := range expectedScripts {
		require.Contains(t, listing, name, "Init script %s not found in container", name)

		rc, err := ctr.CopyFileFromContainer(context.Background(), filepath.Join(dstDir, name))
		require.NoError(t, err, "Failed to copy init script %s from container", name)

		got, err := io.ReadAll(rc)
		require.NoError(t, err, "Failed to read init script %s content", name)

		err = rc.Close()
		require.NoError(t, err, "Failed to close reader for init script %s", name)

		require.Equal(t, want, string(got), "Content of init script %s does not match", name)
	}
}

func requireInitScriptsDoesNotExist(t *testing.T, ctr testcontainers.Container, expectedScripts map[string]string) {
	t.Helper()

	// Sanity check to verify that all scripts are present.
	for filename := range expectedScripts {
		cmd := []string{"sh", "-c", "cd docker-entrypoint-initdb.d && ls -l"}

		exitCode, reader, err := ctr.Exec(context.Background(), cmd)
		require.NoError(t, err)
		require.Zero(t, exitCode, "Expected exit code 0 for command: %v", cmd)

		content, _ := io.ReadAll(reader)
		require.NotContains(t, string(content), filename)
	}
}

func requireContainerLogsNotEmpty(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	logs, err := ctr.Logs(context.Background())
	require.NoError(t, err)

	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	require.NoError(t, err)

	require.NotEmpty(t, logBytes, "Container logs should not be empty")
}

func newAuthFiles(t *testing.T) (string, string, string) {
	t.Helper()

	tmpDir := t.TempDir()

	// Create username and password files.
	usernameFilepath := filepath.Join(tmpDir, "username.txt")

	err := os.WriteFile(usernameFilepath, []byte("file_testuser"), 0o755)
	require.NoError(t, err)

	_, err = os.Stat(usernameFilepath)
	require.NoError(t, err, "Username file should exist")

	// Create the password file.
	passwordFilepath := filepath.Join(tmpDir, "password.txt")

	err = os.WriteFile(passwordFilepath, []byte("file_testpass"), 0o755)
	require.NoError(t, err)

	_, err = os.Stat(passwordFilepath)
	require.NoError(t, err, "Password file should exist")

	return tmpDir, usernameFilepath, passwordFilepath
}
