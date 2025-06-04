package localstack

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort            = 4566
	hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"
	localstackHostEnvVar   = "LOCALSTACK_HOST"
)

var recentVersionTags = []string{
	"latest",
	"s3",
	"s3-latest",
	"stable",
}

func isMinimumVersion(image string, minVersion string) bool {
	parts := strings.Split(strings.Split(image, "@")[0], ":")
	version := parts[len(parts)-1]

	if pos := strings.LastIndexByte(version, '-'); pos >= 0 {
		version = version[0:pos]
	}

	if slices.Contains(recentVersionTags, version) {
		return true
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return semver.IsValid(version) && semver.Compare(version, minVersion) >= 0
}

// WithNetwork creates a network with the given name and attaches the container to it, setting the network alias
// on that network to the given alias.
// Deprecated: use network.WithNetwork or network.WithNewNetwork instead
func WithNetwork(_ string, alias string) testcontainers.CustomizeRequestOption {
	return network.WithNewNetwork(context.Background(), []string{alias})
}

// Deprecated: use Run instead
// RunContainer creates an instance of the LocalStack container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*LocalStackContainer, error) {
	return Run(ctx, "localstack/localstack:1.4.0", opts...)
}

// Run creates an instance of the LocalStack container type
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*LocalStackContainer, error) {
	dockerHost := testcontainers.MustExtractDockerSocket(ctx)

	logger := log.Default()

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(fmt.Sprintf("%d/tcp", defaultPort)),
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{dockerHost + ":/var/run/docker.sock"}
		}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStartupTimeout(120 * time.Second)),
		testcontainers.WithLogger(logger),
	}

	moduleOpts = append(moduleOpts, opts...)

	if !isMinimumVersion(img, "v0.11") {
		return nil, fmt.Errorf("version=%s. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0", img)
	}

	envVar := hostnameExternalEnvVar
	if isMinimumVersion(img, "v2") {
		envVar = localstackHostEnvVar
	}

	moduleOpts = append(moduleOpts, configureDockerHost(logger, envVar))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *LocalStackContainer
	if ctr != nil {
		c = &LocalStackContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// Deprecated: use RunContainer instead
// StartContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func StartContainer(ctx context.Context, overrideReq OverrideContainerRequestOption) (*LocalStackContainer, error) {
	return RunContainer(ctx, overrideReq)
}

func configureDockerHost(logger log.Logger, envVar string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if _, ok := req.Env[envVar]; ok {
			logger.Printf("Setting %s to %s (explicitly as environment variable)\n", envVar, req.Env[envVar])
			return nil
		}

		// if the container is not connected to the default network, use the last network alias in the first network
		// for that we need to check if the container is connected to a network and if it has network aliases
		if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
			alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

			req.Env[envVar] = alias
			logger.Printf("Setting %s to %s (to match last network alias on container with non-default network)\n", envVar, req.Env[envVar])
			return nil
		}

		dockerProvider, err := testcontainers.NewDockerProvider()
		if err != nil {
			return err
		}
		defer dockerProvider.Close()

		daemonHost, err := dockerProvider.DaemonHost(context.Background())
		if err != nil {
			return err
		}

		req.Env[envVar] = daemonHost
		logger.Printf("Setting %s to %s (to match host-routable address for container)\n", envVar, req.Env[envVar])
		return nil
	}
}
