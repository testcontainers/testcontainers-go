//go:build !windows
// +build !windows

package testcontainersdocker

// DockerSocketMountPath is the Docker socket mount path.
var DockerSocketMountPath = "/var/run/docker.sock"

// DockerSocketPath is the path to the Docker socket.
var DockerSocketPath = "/var/run/docker.sock"

// DockerSocketSchema is the Docker socket schema.
var DockerSocketSchema = "unix://"

// DockerSocketPathWithSchema is the path to the Docker socket with the schema.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath
