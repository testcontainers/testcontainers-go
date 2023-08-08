package testcontainersdocker

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitSocketPathsFromDockerClient(t *testing.T) {
	t.Run("Linux", func(t *testing.T) {
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping test on Windows systems")
		}

		initSocketPathsFromDockerClient()

		assert.Equal(t, "unix://", DockerSocketSchema)
		assert.Equal(t, "/var/run/docker.sock", DockerSocketPath)
	})

	t.Run("Windows", func(t *testing.T) {
		// Becuase the init function uses the Docker client to extract the Docker socket,
		// and the client is already compiled for Windows, we can't test this on a non-Windows environment
		if os.Getenv("GOOS") != "windows" {
			t.Skip("Skipping test on non-Windows systems")
		}

		initSocketPathsFromDockerClient()

		assert.Equal(t, "unix://", DockerSocketSchema)
		assert.Equal(t, "//var/run/docker.sock", DockerSocketPath, "The Docker socket path on Windows should have a slash prefix")
	})
}
