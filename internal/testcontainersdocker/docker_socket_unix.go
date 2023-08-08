//go:build !windows
// +build !windows

package testcontainersdocker

// Initialise the Docker socket paths with the Unix socket path
// The value of these variables will be overriden by those obtained
// from the Docker client.
var (
	// DockerSocketPath is the path to the Docker socket.
	DockerSocketPath = "/var/run/docker.sock"

	// DockerSocketPathWithSchema is the path to the Docker socket with the schema.
	DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath
)
