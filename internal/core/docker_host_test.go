package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

// testRemoteHost is a testcontainers host defined in the properties file for testing purposes
var testRemoteHost = TCPSchema + "127.0.0.1:12345"

var (
	originalDockerSocketPath           string
	originalDockerSocketPathWithSchema string
)

var (
	originalDockerSocketOverride string
	tmpSchema                    string
)

func init() {
	originalDockerSocketPath = DockerSocketPath
	originalDockerSocketPathWithSchema = DockerSocketPathWithSchema

	originalDockerSocketOverride = os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

	tmpSchema = DockerSocketSchema
}

var resetSocketOverrideFn = func() {
	os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", originalDockerSocketOverride)
}

func TestExtractDockerHost(t *testing.T) {
	setupDockerHostNotFound(t)
	// do not mess with local .testcontainers.properties
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir) // Windows support

	t.Run("Docker Host as extracted just once", func(t *testing.T) {
		expected := "/path/to/docker.sock"
		t.Setenv("DOCKER_HOST", expected)
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, expected, host)

		t.Setenv("DOCKER_HOST", "/path/to/another/docker.sock")

		host = ExtractDockerHost(context.Background())
		assert.Equal(t, expected, host)
	})

	t.Run("Testcontainers Host is resolved first", func(t *testing.T) {
		t.Setenv("DOCKER_HOST", "/path/to/docker.sock")
		content := "tc.host=" + testRemoteHost

		setupTestcontainersProperties(t, content)

		host := extractDockerHost(context.Background())

		assert.Equal(t, testRemoteHost, host)
	})

	t.Run("Docker Host as environment variable", func(t *testing.T) {
		t.Setenv("DOCKER_HOST", "/path/to/docker.sock")
		host := extractDockerHost(context.Background())

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Malformed Docker Host is passed in context", func(t *testing.T) {
		setupDockerSocketNotFound(t)
		setupRootlessNotFound(t)

		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, DockerHostContextKey, "path-to-docker-sock"))

		assert.Equal(t, DockerSocketPathWithSchema, host)
	})

	t.Run("Malformed Schema Docker Host is passed in context", func(t *testing.T) {
		setupDockerSocketNotFound(t)
		setupRootlessNotFound(t)
		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, DockerHostContextKey, "http://path to docker sock"))

		assert.Equal(t, DockerSocketPathWithSchema, host)
	})

	t.Run("Unix Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := extractDockerHost(context.WithValue(ctx, DockerHostContextKey, DockerSocketSchema+"/this/is/a/sample.sock"))

		assert.Equal(t, "/this/is/a/sample.sock", host)
	})

	t.Run("Unix Docker Host is passed as docker.host", func(t *testing.T) {
		setupDockerSocketNotFound(t)
		setupRootlessNotFound(t)
		content := "docker.host=" + DockerSocketSchema + "/this/is/a/sample.sock"

		setupTestcontainersProperties(t, content)

		host := extractDockerHost(context.Background())

		assert.Equal(t, DockerSocketSchema+"/this/is/a/sample.sock", host)
	})

	t.Run("Default Docker socket", func(t *testing.T) {
		setupRootlessNotFound(t)
		tmpSocket := setupDockerSocket(t)

		host := extractDockerHost(context.Background())

		assert.Equal(t, tmpSocket, host)
	})

	t.Run("Default Docker Host when empty", func(t *testing.T) {
		setupDockerSocketNotFound(t)
		setupRootlessNotFound(t)
		host := extractDockerHost(context.Background())

		assert.Equal(t, DockerSocketPathWithSchema, host)
	})

	t.Run("Extract Docker socket", func(t *testing.T) {
		setupDockerHostNotFound(t)
		t.Cleanup(resetSocketOverrideFn)

		t.Run("Testcontainers host is defined in properties", func(t *testing.T) {
			content := "tc.host=" + testRemoteHost

			setupTestcontainersProperties(t, content)

			socket, err := testcontainersHostFromProperties(context.Background())
			require.NoError(t, err)
			assert.Equal(t, testRemoteHost, socket)
		})

		t.Run("Testcontainers host is not defined in properties", func(t *testing.T) {
			content := "ryuk.disabled=false"

			setupTestcontainersProperties(t, content)

			socket, err := testcontainersHostFromProperties(context.Background())
			require.ErrorIs(t, err, ErrTestcontainersHostNotSetInProperties)
			assert.Empty(t, socket)
		})

		t.Run("DOCKER_HOST is set", func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpSocket := filepath.Join(tmpDir, "docker.sock")
			t.Setenv("DOCKER_HOST", tmpSocket)
			err := createTmpDockerSocket(tmpDir)
			require.NoError(t, err)

			socket, err := dockerHostFromEnv(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("DOCKER_HOST is not set", func(t *testing.T) {
			t.Setenv("DOCKER_HOST", "")

			socket, err := dockerHostFromEnv(context.Background())
			require.ErrorIs(t, err, ErrDockerHostNotSet)
			assert.Empty(t, socket)
		})

		t.Run("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is set", func(t *testing.T) {
			t.Cleanup(resetSocketOverrideFn)

			tmpDir := t.TempDir()
			tmpSocket := filepath.Join(tmpDir, "docker.sock")
			t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", tmpSocket)
			err := createTmpDockerSocket(tmpDir)
			require.NoError(t, err)

			socket, err := dockerSocketOverridePath(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is not set", func(t *testing.T) {
			t.Cleanup(resetSocketOverrideFn)

			os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

			socket, err := dockerSocketOverridePath(context.Background())
			require.ErrorIs(t, err, ErrDockerSocketOverrideNotSet)
			assert.Empty(t, socket)
		})

		t.Run("Context sets the Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerHostFromContext(context.WithValue(ctx, DockerHostContextKey, DockerSocketSchema+"/this/is/a/sample.sock"))
			require.NoError(t, err)
			assert.Equal(t, "/this/is/a/sample.sock", socket)
		})

		t.Run("Context sets a malformed Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerHostFromContext(context.WithValue(ctx, DockerHostContextKey, "path-to-docker-sock"))
			require.Error(t, err)
			assert.Empty(t, socket)
		})

		t.Run("Context sets a malformed schema for the Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerHostFromContext(context.WithValue(ctx, DockerHostContextKey, "http://example.com/docker.sock"))
			require.ErrorIs(t, err, ErrNoUnixSchema)
			assert.Empty(t, socket)
		})

		t.Run("Docker socket exists", func(t *testing.T) {
			tmpSocket := setupDockerSocket(t)

			socket, err := dockerSocketPath(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("Docker host is defined in properties", func(t *testing.T) {
			tmpSocket := "unix:///this/is/a/sample.sock"
			content := "docker.host=" + tmpSocket

			setupTestcontainersProperties(t, content)

			socket, err := dockerHostFromProperties(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("Docker host is not defined in properties", func(t *testing.T) {
			content := "ryuk.disabled=false"

			setupTestcontainersProperties(t, content)

			socket, err := dockerHostFromProperties(context.Background())
			require.ErrorIs(t, err, ErrDockerSocketNotSetInProperties)
			assert.Empty(t, socket)
		})

		t.Run("Docker socket does not exist", func(t *testing.T) {
			setupDockerSocketNotFound(t)

			socket, err := dockerSocketPath(context.Background())
			require.ErrorIs(t, err, ErrSocketNotFoundInPath)
			assert.Empty(t, socket)
		})
	})
}

// mockCli is a mock implementation of client.APIClient, which is handy for simulating
// different operating systems.
type mockCli struct {
	client.APIClient
	OS string
}

// Info returns a mock implementation of types.Info, which is handy for detecting the operating system,
// which is used to determine the default docker socket path.
func (m mockCli) Info(ctx context.Context) (types.Info, error) {
	return types.Info{
		OperatingSystem: m.OS,
	}, nil
}

func TestExtractDockerSocketFromClient(t *testing.T) {
	setupDockerHostNotFound(t)

	t.Run("Docker socket from Testcontainers host defined in properties", func(t *testing.T) {
		content := "tc.host=" + testRemoteHost

		setupTestcontainersProperties(t, content)

		socket := extractDockerSocketFromClient(context.Background(), mockCli{OS: "foo"})
		assert.Equal(t, DockerSocketPath, socket)
	})

	t.Run("Docker socket from Testcontainers host takes precedence over TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", func(t *testing.T) {
		content := "tc.host=" + testRemoteHost

		setupTestcontainersProperties(t, content)

		t.Cleanup(resetSocketOverrideFn)
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/path/to/docker.sock")

		socket := extractDockerSocketFromClient(context.Background(), mockCli{OS: "foo"})
		assert.Equal(t, DockerSocketPath, socket)
	})

	t.Run("Docker Socket as Testcontainers environment variable", func(t *testing.T) {
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/path/to/docker.sock")
		host := extractDockerSocketFromClient(context.Background(), mockCli{OS: "foo"})

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Docker Socket as Testcontainers environment variable, removes prefixes", func(t *testing.T) {
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", DockerSocketSchema+"/path/to/docker.sock")
		host := extractDockerSocketFromClient(context.Background(), mockCli{OS: "foo"})
		assert.Equal(t, "/path/to/docker.sock", host)

		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", testRemoteHost)
		host = extractDockerSocketFromClient(context.Background(), mockCli{OS: "foo"})
		assert.Equal(t, DockerSocketPath, host)
	})

	t.Run("Unix Docker Socket is passed as DOCKER_HOST variable (Docker Desktop on non-Windows)", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Skip for Windows")
		}

		t.Setenv("GOOS", "linux")
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		ctx := context.Background()
		os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		t.Setenv("DOCKER_HOST", DockerSocketSchema+"/this/is/a/sample.sock")

		socket := extractDockerSocketFromClient(ctx, mockCli{OS: "Docker Desktop"})

		assert.Equal(t, DockerSocketPath, socket)
	})

	t.Run("Unix Docker Socket is passed as DOCKER_HOST variable (Docker Desktop for Windows)", func(t *testing.T) {
		t.Setenv("GOOS", "windows")
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		ctx := context.Background()
		os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		t.Setenv("DOCKER_HOST", DockerSocketSchema+"/this/is/a/sample.sock")

		socket := extractDockerSocketFromClient(ctx, mockCli{OS: "Docker Desktop"})

		assert.Equal(t, WindowsDockerSocketPath, socket)
	})

	t.Run("Unix Docker Socket is passed as DOCKER_HOST variable (Not Docker Desktop)", func(t *testing.T) {
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		ctx := context.Background()
		os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		t.Setenv("DOCKER_HOST", DockerSocketSchema+"/this/is/a/sample.sock")

		socket := extractDockerSocketFromClient(ctx, mockCli{OS: "Ubuntu"})

		assert.Equal(t, "/this/is/a/sample.sock", socket)
	})

	t.Run("Unix Docker Socket is passed as DOCKER_HOST variable (Not Docker Desktop), removes prefixes", func(t *testing.T) {
		setupTestcontainersProperties(t, "")

		t.Cleanup(resetSocketOverrideFn)

		ctx := context.Background()
		os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

		t.Setenv("DOCKER_HOST", DockerSocketSchema+"/this/is/a/sample.sock")
		socket := extractDockerSocketFromClient(ctx, mockCli{OS: "Ubuntu"})
		assert.Equal(t, "/this/is/a/sample.sock", socket)

		t.Setenv("DOCKER_HOST", testRemoteHost)
		socket = extractDockerSocketFromClient(ctx, mockCli{OS: "Ubuntu"})
		assert.Equal(t, DockerSocketPath, socket)
	})

	t.Run("Unix Docker Socket is passed as docker.host property", func(t *testing.T) {
		content := "docker.host=" + DockerSocketSchema + "/this/is/a/sample.sock"
		setupTestcontainersProperties(t, content)
		setupDockerSocketNotFound(t)

		t.Cleanup(resetSocketOverrideFn)

		ctx := context.Background()
		os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		os.Unsetenv("DOCKER_HOST")

		socket := extractDockerSocketFromClient(ctx, mockCli{OS: "Ubuntu"})

		assert.Equal(t, "/this/is/a/sample.sock", socket)
	})
}

func TestInAContainer(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		assert.False(t, inAContainer(filepath.Join(tmpDir, ".dockerenv-a")))
	})

	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		f := filepath.Join(tmpDir, ".dockerenv-b")

		testFile, err := os.Create(f)
		require.NoError(t, err)
		defer testFile.Close()

		assert.True(t, inAContainer(f))
	})
}

func createTmpDir(dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	return nil
}

func createTmpDockerSocket(parent string) error {
	socketPath := filepath.Join(parent, "docker.sock")
	err := os.MkdirAll(filepath.Dir(socketPath), 0o755)
	if err != nil {
		return err
	}

	f, err := os.Create(socketPath)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

// setupDockerHostNotFound sets up the environment for the test case where the DOCKER_HOST environment variable is
// already set (e.g. rootless docker) therefore we need to unset it before the test
func setupDockerHostNotFound(t *testing.T) {
	t.Setenv("DOCKER_HOST", "")
}

func setupDockerSocket(t *testing.T) string {
	t.Cleanup(func() {
		DockerSocketPath = originalDockerSocketPath
		DockerSocketPathWithSchema = originalDockerSocketPathWithSchema
	})

	tmpDir := t.TempDir()
	tmpSocket := filepath.Join(tmpDir, "docker.sock")
	err := createTmpDockerSocket(filepath.Dir(tmpSocket))
	require.NoError(t, err)

	DockerSocketPath = tmpSocket
	DockerSocketPathWithSchema = tmpSchema + tmpSocket

	return tmpSchema + tmpSocket
}

func setupDockerSocketNotFound(t *testing.T) {
	t.Cleanup(func() {
		DockerSocketPath = originalDockerSocketPath
		DockerSocketPathWithSchema = originalDockerSocketPathWithSchema
	})

	tmpDir := t.TempDir()
	tmpSocket := filepath.Join(tmpDir, "docker.sock")

	DockerSocketPath = tmpSocket
}

func setupTestcontainersProperties(t *testing.T, content string) {
	t.Cleanup(func() {
		// reset the properties file after the test
		config.Reset()
	})

	config.Reset()

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	err := createTmpDir(homeDir)
	require.NoError(t, err)
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir) // Windows support

	if err := os.WriteFile(filepath.Join(homeDir, ".testcontainers.properties"), []byte(content), 0o600); err != nil {
		t.Errorf("Failed to create the file: %v", err)
		return
	}
}
