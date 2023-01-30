package testcontainersdocker

import (
	"context"
	"net/url"
	"os"
)

type dockerHostContext string

var DockerHostContextKey = dockerHostContext("docker_host")

// Extracts the docker host from the context, or returns the default value
func ExtractDockerHost(ctx context.Context) (dockerHostPath string) {
	if dockerHostPath = os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"); dockerHostPath != "" {
		return dockerHostPath
	}

	dockerHostPath = "/var/run/docker.sock"

	var hostRawURL string
	if h, ok := ctx.Value(DockerHostContextKey).(string); !ok || h == "" {
		return dockerHostPath
	} else {
		hostRawURL = h
	}
	var hostURL *url.URL
	if u, err := url.Parse(hostRawURL); err != nil {
		return dockerHostPath
	} else {
		hostURL = u
	}

	switch hostURL.Scheme {
	case "unix":
		return hostURL.Path
	default:
		return dockerHostPath
	}
}
