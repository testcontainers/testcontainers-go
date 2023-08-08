package testcontainersdocker

import (
	"net/url"
	"strings"

	"github.com/docker/docker/client"
)

// DockerSocketSchema is the unix schema.
var DockerSocketSchema = "unix://"

// DockerSocketPath is the path to the docker socket under unix systems.
var DockerSocketPath = "/var/run/docker.sock"

// DockerSocketPathWithSchema is the path to the docker socket under unix systems with the unix schema.
var DockerSocketPathWithSchema = DockerSocketSchema + DockerSocketPath

// TCPSchema is the tcp schema.
var TCPSchema = "tcp://"

func init() {
	initSocketPaths()
}

func initSocketPaths() {
	const DefaultDockerHost = client.DefaultDockerHost

	u, err := url.Parse(DefaultDockerHost)
	if err != nil {
		// unsupported default host specified by the docker client package,
		// so revert to the default unix docker socket path
		return
	}

	var schema string
	var socketPath string

	switch u.Scheme {
	case "unix", "npipe":
		schema = u.Scheme + "://"
		socketPath = u.Path
		if !strings.HasPrefix(socketPath, "/") {
			// seeing as the code in this module depends on DockerSocketPath having
			// a slash (`/`) prefix, we add it here if it is missing.
			// for the known environments, we do not foresee how the socket-path
			// should miss the slash, however this extra if-condition is worth to
			// save future pain from innocent users.
			socketPath = "/" + socketPath
		}

		if u.Scheme == "npipe" {
			// the docker client package uses the npipe schema for windows
			// docker sockets, so we need to replace it with the unix schema
			schema = DockerSocketSchema
			// the docker socket path on Windows using Linux containers
			// prepends the slash to the unix socket path
			socketPath = "/" + DockerSocketPath
		}
	}

	DockerSocketSchema = schema
	DockerSocketPath = socketPath
	DockerSocketPathWithSchema = schema + socketPath
}
