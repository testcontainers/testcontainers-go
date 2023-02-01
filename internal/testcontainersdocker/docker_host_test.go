package testcontainersdocker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractDockerHost(t *testing.T) {
	t.Run("Docker Host as environment variable", func(t *testing.T) {
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/path/to/docker.sock")
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, "/path/to/docker.sock", host)
	})

	t.Run("Default Docker Host", func(t *testing.T) {
		host := ExtractDockerHost(context.Background())

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Malformed Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "path-to-docker-sock"))

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Malformed Schema Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "http://path to docker sock"))

		assert.Equal(t, "/var/run/docker.sock", host)
	})

	t.Run("Unix Docker Host is passed in context", func(t *testing.T) {
		ctx := context.Background()

		host := ExtractDockerHost(context.WithValue(ctx, DockerHostContextKey, "unix:///this/is/a/sample.sock"))

		assert.Equal(t, "/this/is/a/sample.sock", host)
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
