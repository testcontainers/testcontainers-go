package testcontainersdocker

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type dockerHostContext string

var DockerHostContextKey = dockerHostContext("docker_host")

const DefaultDockerSocketPath = "/var/run/docker.sock"

var (
	ErrDockerSocketOverrideNotSet = errors.New("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is not set")
	ErrSocketNotFound             = errors.New("socket not found")
)

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
func ExtractDockerHost(ctx context.Context) string {
	socketPathFns := []func(context.Context) (string, error){
		dockerSocketOverridePath,
		extractDockerSocketPath,
	}

	outerErr := ErrSocketNotFound
	for _, socketPathFn := range socketPathFns {
		socketPath, err := socketPathFn(ctx)
		if err != nil {
			outerErr = fmt.Errorf("%w: %v", outerErr, err)
			continue
		}

		return socketPath
	}

	return DefaultDockerSocketPath
}

func dockerSocketOverridePath(ctx context.Context) (string, error) {
	if dockerHostPath, exists := os.LookupEnv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"); exists {
		return dockerHostPath, nil
	}

	return "", ErrDockerSocketOverrideNotSet
}

// extractDockerSocketPath returns if the path to the Docker socket exists.
func extractDockerSocketPath(ctx context.Context) (string, error) {
	dockerHostPath := DefaultDockerSocketPath

	var hostRawURL string
	if h, ok := ctx.Value(DockerHostContextKey).(string); !ok || h == "" {
		return dockerHostPath, nil
	} else {
		hostRawURL = h
	}
	var hostURL *url.URL
	if u, err := url.Parse(hostRawURL); err != nil {
		return dockerHostPath, nil
	} else {
		hostURL = u
	}

	switch hostURL.Scheme {
	case "unix":
		return hostURL.Path, nil
	default:
		return dockerHostPath, nil
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
