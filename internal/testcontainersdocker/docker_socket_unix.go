//go:build !windows
// +build !windows

package testcontainersdocker

// DockerSocketPathWithSchema is the path to the Docker socket with the schema.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath
