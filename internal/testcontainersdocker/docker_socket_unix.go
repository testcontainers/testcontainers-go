//go:build !windows
// +build !windows

package testcontainersdocker

// DockerSocketPath is the path to the Docker socket.
var DockerSocketPath = "/var/run/docker.sock"

// DockerSocketSchema is the Docker socket schema.
var DockerSocketSchema = "unix://"
