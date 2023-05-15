package testcontainersdocker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/testcontainers/testcontainers-go/internal/config"
)

type dockerHostContext string

var DockerHostContextKey = dockerHostContext("docker_host")

var (
	ErrDockerHostNotSet               = errors.New("DOCKER_HOST is not set")
	ErrDockerSocketOverrideNotSet     = errors.New("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE is not set")
	ErrDockerSocketNotSetInContext    = errors.New("socket not set in context")
	ErrDockerSocketNotSetInProperties = errors.New("socket not set in ~/.testcontainers.properties")
	ErrNoUnixSchema                   = errors.New("URL schema is not unix")
	ErrSocketNotFound                 = errors.New("socket not found")
	ErrSocketNotFoundInPath           = errors.New("docker socket not found in " + DockerSocketPath)
)

var dockerHostCache string
var dockerHostOnce sync.Once

var mountDockerHostCache string
var mountDockerHostOnce sync.Once

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

// ExtractDockerHost Extracts the docker host from the different alternatives, caching the result to avoid unnecessary
// calculations.
// The possible alternatives are:
// 1. DOCKER_HOST environment variable
// 2. TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE environment variable
// 3. Docker socket path from context
// 4. Docker socket path from the default docker socket path
// 5. Docker socket path from the "docker.host" property in the ~/.testcontainers.properties file
// 6. Rootless docker socket path
// If none of the above alternatives are found, an empty string is returned
func ExtractDockerHost(ctx context.Context) string {
	dockerHostOnce.Do(func() {
		dockerHostCache = extractDockerHost(ctx)
	})

	return dockerHostCache
}

// ExtractMountDockerHost Extracts the docker host from the different alternatives, removing the socket schema
func ExtractMountDockerHost(ctx context.Context) string {
	mountDockerHostOnce.Do(func() {
		mountDockerHostCache = extractMountDockerHost(ctx)
	})

	return mountDockerHostCache
}

// extractDockerHost Extracts the docker host from the different alternatives, without caching the result.
// This internal method is handy for testing purposes.
func extractDockerHost(ctx context.Context) string {
	socketPathFns := []func(context.Context) (string, error){
		dockerHostFromEnv,
		dockerSocketOverridePath,
		dockerSocketFromContext,
		dockerSocketPath,
		dockerSocketFromProperties,
		rootlessDockerSocketPath,
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

	return ""
}

// extractDockerHost Extracts the docker host from the different alternatives, without caching the result.
// This internal method is handy for testing purposes.
func extractMountDockerHost(ctx context.Context) string {
	dockerHost := extractDockerHost(ctx)

	return strings.Replace(dockerHost, DockerSocketSchema, "", 1)
}

// dockerHostFromEnv returns the docker host from the DOCKER_HOST environment variable, if it's not empty
func dockerHostFromEnv(ctx context.Context) (string, error) {
	if dockerHostPath := os.Getenv("DOCKER_HOST"); dockerHostPath != "" {
		return dockerHostPath, nil
	}

	return "", ErrDockerHostNotSet
}

func dockerSocketFromContext(ctx context.Context) (string, error) {
	if socketPath, ok := ctx.Value(DockerHostContextKey).(string); ok && socketPath != "" {
		parsed, err := parseURL(socketPath)
		if err != nil {
			return "", err
		}

		return parsed, nil
	}

	return "", ErrDockerSocketNotSetInContext
}

func dockerSocketFromProperties(ctx context.Context) (string, error) {
	cfg := config.Read(ctx)
	socketPath := cfg.Host
	if socketPath != "" {
		parsed, err := parseURL(socketPath)
		if err != nil {
			return "", err
		}

		return parsed, nil
	}

	return "", ErrDockerSocketNotSetInProperties
}

func dockerSocketOverridePath(ctx context.Context) (string, error) {
	if dockerHostPath, exists := os.LookupEnv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE"); exists {
		return dockerHostPath, nil
	}

	return "", ErrDockerSocketOverrideNotSet
}

func dockerSocketPath(ctx context.Context) (string, error) {
	if fileExists(DockerSocketPath) {
		return DockerSocketPathWithSchema, nil
	}

	return "", ErrSocketNotFoundInPath
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
