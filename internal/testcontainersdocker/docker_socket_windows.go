//go:build windows
// +build windows

package testcontainersdocker

// DockerSocketSchema is the Docker socket schema on Windows
var DockerSocketSchema = "npipe://"

// DockerSocketPath The socket path for Windows contains two slashes, exactly the same as the Docker Desktop socket path.
var DockerSocketPath = "//./pipe/docker_engine"
