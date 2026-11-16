package atlaslocal_test

import (
	"context"
	"io"
	"math/rand/v2"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/mongodb/atlaslocal"
	"github.com/testcontainers/testcontainers-go/modules/mongodb/internal/mongotest"
)

const latestImage = "mongodb/mongodb-atlas-local:latest"

func TestMongoDBAtlasLocal(t *testing.T) {
	ctx := context.Background()

	ctr, err := atlaslocal.Run(ctx, latestImage)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	client := newMongoClient(t, ctx, ctr)
	var ping int
	err = client.Run(ctx, `return commandWorked(db.getSiblingDB("admin").runCommand({ping:1})).ok;`, &ping)
	require.NoError(t, err)
	require.Equal(t, 1, ping)
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

			client := newMongoClient(t, context.Background(), ctr)
			// Execute the original authenticated write through the generated endpoint.
			var acknowledged bool
			err = client.Run(context.Background(), `return acknowledged(db.getSiblingDB('test').getCollection('foo').insertOne({test:'value'})).acknowledged;`, &acknowledged)
			require.NoError(t, err, "Failed to insert document with authentication")
			require.True(t, acknowledged)
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

	client := newMongoClient(t, context.Background(), ctr)
	seed := `{"_id":{"$numberInt":"1"},"seeded":true}`
	var raw string
	err = client.Run(context.Background(), `
const doc = db.getSiblingDB('mydb').getCollection('foo').findOne(EJSON.parse('{"_id":{"$numberInt":"1"},"seeded":true}', {relaxed:false}), {}, {promoteValues:false, promoteLongs:false});
if (doc === null) throw new Error('seeded document not found');
return EJSON.stringify(doc, null, 0, {relaxed:false});
`, &raw)
	require.NoError(t, err)
	doc, err := mongotest.CanonicalDocument(raw)
	require.NoError(t, err, "Failed to decode seeded document")
	require.Equal(t, seed, doc, "Seeded document does not match expected values")
}

func TestWithInitScripts(t *testing.T) {
	seed1 := `{"_id":{"$numberInt":"1"},"seeded":true}`
	seed2 := `{"_id":{"$numberInt":"2"},"seeded":true}`

	cases := []struct {
		name        string
		initScripts map[string]string // filename -> content
		want        []string
	}{
		{
			name:        "no scripts",
			initScripts: map[string]string{},
			want:        []string{},
		},
		{
			name: "single shell script",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
			},
			want: []string{seed1},
		},
		{
			name: "single js script",
			initScripts: map[string]string{
				"01-seed.js": `db.foo.insertOne({ _id: 1, seeded: true });`,
			},
			want: []string{seed1},
		},
		{
			name: "mixed shell and js scripts",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
				"02-seed.js": `db.foo.insertOne({ _id: 2, seeded: true });`,
			},
			want: []string{seed1, seed2},
		},
		{
			name: "mixed ordered shell",
			initScripts: map[string]string{
				"01-seed.sh": `mongosh --eval 'db.foo.insertOne({ _id: 1, seeded: true })'`,
				"02-seed.sh": `mongosh --eval 'db.foo.deleteOne({ _id: 1, seeded: true })'`,
			},
			want: []string{},
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

			client := newMongoClient(t, context.Background(), ctr)
			var raw []string
			err = client.Run(context.Background(), `
const cursor = db.getSiblingDB('test').getCollection('foo').find({}, {}, {promoteValues:false, promoteLongs:false});
try { return cursor.toArray().map(doc => EJSON.stringify(doc, null, 0, {relaxed:false})); }
finally { cursor.close(); }
`, &raw)
			require.NoError(t, err)
			results := make([]string, 0, len(raw))
			for _, doc := range raw {
				canonical, err := mongotest.CanonicalDocument(doc)
				require.NoError(t, err)
				results = append(results, canonical)
			}

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
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			csRaw, err := ctr.ConnectionString(context.Background())
			require.NoError(t, err)

			connString, err := url.Parse(csRaw)
			require.NoError(t, err, "Failed to parse connection string")

			require.Equal(t, "mongodb", connString.Scheme)
			host, port, err := net.SplitHostPort(connString.Host)
			require.NoError(t, err, "Invalid host:port")
			require.NotEmpty(t, host)
			require.NotEmpty(t, port)

			require.Equal(t, tc.wantUsername, connString.User.Username())
			password, _ := connString.User.Password()
			require.Equal(t, tc.wantPassword, password)
			require.Equal(t, tc.wantDatabase, strings.TrimPrefix(connString.Path, "/"))
			require.Equal(t, "true", connString.Query().Get("directConnection"))
		})
	}
}

// Test Helper Functions

func requireEnvVar(t *testing.T, ctr testcontainers.Container, envVarName, expected string) {
	t.Helper()

	// testcontainers-go's Exec() returns a multiplexed stream in the same format
	// used by the Docker API. Each frame is prefixed with an 8-byte header.
	exitCode, reader, err := ctr.Exec(context.Background(), []string{"sh", "-c", "echo $" + envVarName}, exec.Multiplexed())
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	outBytes, err := io.ReadAll(reader)
	require.NoError(t, err)

	out := strings.TrimSpace(string(outBytes))
	require.Equal(t, expected, out, "DO_NOT_TRACK env var value mismatch")
}

func requireMongotLogs(t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	// Pull the log file and require non-empty.
	reader, err := ctr.(*atlaslocal.Container).ReadMongotLogs(context.Background())
	require.NoError(t, err)
	defer reader.Close()

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

// executeAggregation preserves the original collection/index/write/search traffic.
// Iterating the lazy shell cursor forces the aggregation command to execute.
func executeAggregation(t *testing.T, ctr testcontainers.Container) {
	t.Helper()
	client := newMongoClient(t, context.Background(), ctr)
	var result struct {
		Collection, Index, Aggregation bool
		Acknowledged                   bool
	}
	err := client.Run(context.Background(), `
const database = db.getSiblingDB('test');
commandWorked(database.createCollection('search'));
commandWorked(database.runCommand({createSearchIndexes:'search', indexes:[{name:'test_search_index', definition:{mappings:{dynamic:true}}}]}));
const coll = database.getCollection('search');
const inserted = acknowledged(coll.insertOne({txt:'hello'}));
const cursor = coll.aggregate([{$search:{text:{query:'hello', path:'txt'}}}], {maxTimeMS:90000});
try { cursor.toArray(); }
finally { cursor.close(); }
return {collection:true, index:true, acknowledged:inserted.acknowledged, aggregation:true};
`, &result)
	require.NoError(t, err)
	require.True(t, result.Collection)
	require.True(t, result.Index)
	require.True(t, result.Acknowledged)
	require.True(t, result.Aggregation)
}

func newMongoClient(t *testing.T, ctx context.Context, ctr testcontainers.Container) *mongotest.Shell {
	t.Helper()
	connString, err := ctr.(*atlaslocal.Container).ConnectionString(ctx)
	require.NoError(t, err)
	return mongotest.New(t, ctx, connString)
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

	exit, r, err := ctr.Exec(context.Background(), []string{"sh", "-lc", "ls -l " + dstDir}, exec.Multiplexed())
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
		cmd := []string{"sh", "-lc", "ls -1 /docker-entrypoint-initdb.d 2>/dev/null || true"}
		_, reader, err := ctr.Exec(context.Background(), cmd, exec.Multiplexed())
		require.NoError(t, err)

		raw, err := io.ReadAll(reader)
		require.NoError(t, err)

		require.NotContains(t, string(raw), filename)
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

	err := os.WriteFile(usernameFilepath, []byte("file_testuser"), 0o600)
	require.NoError(t, err)

	_, err = os.Stat(usernameFilepath)
	require.NoError(t, err, "Username file should exist")

	// Create the password file.
	passwordFilepath := filepath.Join(tmpDir, "password.txt")

	err = os.WriteFile(passwordFilepath, []byte("file_testpass"), 0o600)
	require.NoError(t, err)

	_, err = os.Stat(passwordFilepath)
	require.NoError(t, err, "Password file should exist")

	return tmpDir, usernameFilepath, passwordFilepath
}
