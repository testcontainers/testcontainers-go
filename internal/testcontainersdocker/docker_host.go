package testcontainersdocker

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type dockerHostContext string

var DockerHostContextKey = dockerHostContext("docker_host")

// deprecated
// see https://github.com/testcontainers/testcontainers-java/blob/main/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L46
func DefaultGatewayIP() (string, error) {
	// see https://github.com/testcontainers/testcontainers-java/blob/3ad8d80e2484864e554744a4800a81f6b7982168/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L27
	cmd := exec.Command("sh", "-c", "ip route|awk '/default/ { print $3 }'")
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.New("failed to detect docker host")
	}
	ip := strings.TrimSpace(string(stdout))
	if len(ip) == 0 {
		return "", errors.New("failed to parse default gateway IP")
	}
	return ip, nil
}

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

// InAContainer returns true if the code is running inside a container
// See https://github.com/docker/docker/blob/a9fa38b1edf30b23cae3eade0be48b3d4b1de14b/daemon/initlayer/setup_unix.go#L25
func InAContainer() bool {
	return inAContainer("/.dockerenv")
}

func inAContainer(path string) bool {
	// see https://github.com/testcontainers/testcontainers-java/blob/3ad8d80e2484864e554744a4800a81f6b7982168/core/src/main/java/org/testcontainers/dockerclient/DockerClientConfigUtils.java#L15
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
