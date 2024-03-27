package core

import (
	"context"
	"net/url"
	"strings"

	"github.com/docker/docker/client"
)

// DockerSocketSchema is the unix schema.
var DockerSocketSchema = "unix://"

// DockerSocketPath is the path to the docker socket under unix systems.
var DockerSocketPath = "/var/run/docker.sock"

// TCPSchema is the tcp schema.
const TCPSchema = "tcp://"

// windowsDockerSocketPath is the path to the docker socket under windows systems and Linux containers.
const windowsDockerSocketPath = "//var/run/docker.sock"

func init() {
	const DefaultDockerHost = client.DefaultDockerHost

	u, err := url.Parse(DefaultDockerHost)
	if err != nil {
		// unsupported default host specified by the docker client package,
		// so revert to the default unix docker socket path
		return
	}

	switch u.Scheme {
	case "unix", "npipe":
		DockerSocketSchema = u.Scheme + "://"
		DockerSocketPath = u.Path
		if !strings.HasPrefix(DockerSocketPath, "/") {
			// seeing as the code in this module depends on DockerSocketPath having
			// a slash (`/`) prefix, we add it here if it is missing.
			// for the known environments, we do not foresee how the socket-path
			// should miss the slash, however this extra if-condition is worth to
			// save future pain from innocent users.
			DockerSocketPath = "/" + DockerSocketPath
		}
	}
}

// checkDefaultDockerSocket checks the docker socket path and returns the correct path depending on the Docker client configuration,
// the operating system, and the Docker info.
// It will use the Docker client infrastructure to get the correct path:
// - If the Docker client is running in a local Docker host, it will return "/var/run/docker.sock" on Unix, or "//./pipe/docker_engine" on Windows.
// - If the Docker client is running in Docker Desktop, it will return "/var/run/docker.sock" on Unix, or "//./pipe/docker_engine" on Windows.
// - If the Docker client is running in a remote Docker host, it will return "/var/run/docker.sock" on Unix, or "//var/run/docker.sock" on Windows.
// - If the Docker client is running in a rootless Docker, it will return the proper path depending on the rootless Docker configuration. Not available for windows.
// If the Docker info cannot be retrieved, the program will panic.
// This internal method is handy for testing purposes, passing a mock type simulating the desired behaviour.
func checkDefaultDockerSocket(ctx context.Context, cli client.APIClient, socket string) string {
	info, err := cli.Info(ctx)
	if err != nil {
		panic(err) // Docker Info is required to get the client socket
	}

	// this default path will come from the default docker client, which for
	// unix systems is "/var/run/docker.sock", and for Windows is "//./pipe/docker_engine"
	defaultDockerSocketPath := "/var/run/docker.sock"
	defaultSocketSchema := "unix://"
	defaultRemoteDockerSocketPath := defaultDockerSocketPath
	defaultRootlessDockerSocketPath := defaultDockerSocketPath

	if IsWindows() {
		// the path to the docker socket under windows systems and Windows containers.
		defaultDockerSocketPath = "//./pipe/docker_engine"
		defaultSocketSchema = "npipe://"
		defaultRemoteDockerSocketPath = "//var/run/docker.sock"
		defaultRootlessDockerSocketPath = "//var/run/docker.sock"
	} else {
		// rootless docker socket path is not supported on Windows
		defaultRootlessDockerSocketPath, err = rootlessDockerSocketPath(ctx)
		if err != nil {
			defaultRootlessDockerSocketPath = defaultDockerSocketPath
		}

		// rootless is enabled, so we need to use the default docker paths for rootless docker.
		if defaultDockerSocketPath != defaultRootlessDockerSocketPath {
			defaultDockerSocketPath = defaultRootlessDockerSocketPath
			defaultRemoteDockerSocketPath = defaultRootlessDockerSocketPath
		}
	}

	if info.OperatingSystem == "Docker Desktop" {
		// Because Docker Desktop runs in a VM, we need to use the default docker path for rootless docker.
		// For Windows it will be "//var/run/docker.sock" and for Unix it will be "/var/run/docker.sock"
		if strings.HasPrefix(defaultRootlessDockerSocketPath, defaultSocketSchema) {
			return strings.Replace(defaultRootlessDockerSocketPath, defaultSocketSchema, "", 1)
		}

		return defaultRootlessDockerSocketPath
	}

	if info.OSType == "linux" {
		// we are using a remote Docker host, so we need to use the default docker path for rootless docker.
		// For Windows it will be "//var/run/docker.sock" and for Unix it will be "/var/run/docker.sock"
		if strings.HasPrefix(defaultRemoteDockerSocketPath, defaultSocketSchema) {
			return strings.Replace(defaultRemoteDockerSocketPath, defaultSocketSchema, "", 1)
		}

		return defaultRemoteDockerSocketPath
	}

	// check that the socket is not a tcp or unix socket,
	// including potential remote Docker hosts and rootless Docker.

	// this use case will cover the case when the docker host is a tcp socket
	if strings.HasPrefix(socket, TCPSchema) {
		return defaultDockerSocketPath
	}

	// this use case will cover the case when the docker host is a unix or npipe socket
	if strings.HasPrefix(socket, defaultSocketSchema) {
		return strings.Replace(socket, defaultSocketSchema, "", 1)
	}

	return socket
}
