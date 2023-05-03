package testcontainersdocker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ExtractDockerHost(t *testing.T) {
	t.Run("Docker Host as environment variable", func(t *testing.T) {
		t.Setenv("DOCKER_HOST", "/path/to/docker.sock")
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Docker Host as environment variable", func(t *testing.T) {
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/path/to/docker.sock")
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Default Docker Host", func(t *testing.T) {
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, DefaultDockerSocketPath, host)
	})

	t.Run("Malformed Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "path-to-docker-sock"))

		assert.Equal(t, DefaultDockerSocketPath, host)
	})

	t.Run("Malformed Schema Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "http://path to docker sock"))

		assert.Equal(t, DefaultDockerSocketPath, host)
	})

	t.Run("Unix Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "unix:///this/is/a/sample.sock"))

		assert.Equal(t, "/this/is/a/sample.sock", host)
	})

	t.Run("Extract Docker socket", func(t *testing.T) {
		originalDockerSocketOverride := os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")
		defer func() {
			os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", originalDockerSocketOverride)
		}()

		t.Run("DOCKER_HOST is set", func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpSocket := filepath.Join(tmpDir, "docker.sock")
			t.Setenv("DOCKER_HOST", tmpSocket)
			createTmpDockerSocket(tmpDir)

			socket, err := dockerHostFromEnv(context.Background())
			require.Nil(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("DOCKER_HOST is not set", func(t *testing.T) {
			t.Setenv("DOCKER_HOST", "")

			socket, err := dockerHostFromEnv(context.Background())
			require.ErrorIs(t, err, ErrDockerHostNotSet)
			assert.Empty(t, socket)
		})

		t.Run("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is set", func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpSocket := filepath.Join(tmpDir, "docker.sock")
			t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", tmpSocket)
			createTmpDockerSocket(tmpDir)

			socket, err := dockerSocketOverridePath(context.Background())
			require.Nil(t, err)
			assert.Equal(t, tmpSocket, socket)
		})

		t.Run("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is not set", func(t *testing.T) {
			os.Unsetenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")

			socket, err := dockerSocketOverridePath(context.Background())
			require.ErrorIs(t, err, ErrDockerSocketOverrideNotSet)
			assert.Empty(t, socket)
		})

		t.Run("Context sets the Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerSocketFromContext(context.WithValue(ctx, DockerHostContextKey, "unix:///this/is/a/sample.sock"))
			require.Nil(t, err)
			assert.Equal(t, "/this/is/a/sample.sock", socket)
		})

		t.Run("Context sets a malformed Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerSocketFromContext(context.WithValue(ctx, DockerHostContextKey, "path-to-docker-sock"))
			require.Error(t, err)
			assert.Empty(t, socket)
		})

		t.Run("Context sets a malformed schema for the Docker socket", func(t *testing.T) {
			ctx := context.Background()

			socket, err := dockerSocketFromContext(context.WithValue(ctx, DockerHostContextKey, "http://example.com/docker.sock"))
			require.ErrorIs(t, err, ErrNoUnixSchema)
			assert.Empty(t, socket)
		})
	})
}

func TestInAContainer(t *testing.T) {
	const dockerenvName = ".dockerenv"

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		assert.False(t, inAContainer(filepath.Join(tmpDir, dockerenvName)))
	})

	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		f := filepath.Join(tmpDir, dockerenvName)

		_, err := os.Create(f)
		assert.NoError(t, err)
		assert.True(t, inAContainer(f))
	})
}

func createTmpDir(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	return nil
}

func createTmpDockerSocket(parent string) error {
	socketPath := filepath.Join(parent, "docker.sock")
	err := os.MkdirAll(filepath.Dir(socketPath), 0755)
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
