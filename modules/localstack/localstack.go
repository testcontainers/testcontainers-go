package localstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort            = 4566
	hostnameExternalEnvVar = "HOSTNAME_EXTERNAL"
	localstackHostEnvVar   = "LOCALSTACK_HOST"
)

func isLegacyMode(image string) bool {
	parts := strings.Split(image, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return false
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v0.11") < 0 // version < v0.11
	}

	return true
}

func isVersion2(image string) bool {
	parts := strings.Split(image, ":")
	version := parts[len(parts)-1]

	if version == "latest" {
		return true
	}

	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}

	if semver.IsValid(version) {
		return semver.Compare(version, "v2.0") > 0 // version >= v2.0
	}

	return true
}

// WithNetwork creates a network with the given name and attaches the container to it, setting the network alias
// on that network to the given alias.
// Deprecated: use network.WithNetwork or network.WithNewNetwork instead
func WithNetwork(networkName string, alias string) testcontainers.CustomizeRequestOption {
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

	req := testcontainers.ContainerRequest{
		Image:        img,
		WaitingFor:   wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStartupTimeout(120 * time.Second),
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", defaultPort)},
		Env:          map[string]string{},
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{fmt.Sprintf("%s:/var/run/docker.sock", dockerHost)}
		},
	}

	localStackReq := LocalStackContainerRequest{
		GenericContainerRequest: testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Logger:           testcontainers.Logger,
			Started:          true,
		},
	}

	for _, opt := range opts {
		if err := opt.Customize(&localStackReq.GenericContainerRequest); err != nil {
			return nil, err
		}
	}

	if isLegacyMode(localStackReq.Image) {
		return nil, fmt.Errorf("version=%s. Testcontainers for Go does not support running LocalStack in legacy mode. Please use a version >= 0.11.0", localStackReq.Image)
	}

	envVar := hostnameExternalEnvVar
	if isVersion2(localStackReq.Image) {
		envVar = localstackHostEnvVar
	}

	hostnameExternalReason, err := configureDockerHost(&localStackReq, envVar)
	if err != nil {
		return nil, err
	}
	localStackReq.GenericContainerRequest.Logger.Printf("Setting %s to %s (%s)\n", envVar, req.Env[envVar], hostnameExternalReason)

	container, err := testcontainers.GenericContainer(ctx, localStackReq.GenericContainerRequest)
	var c *LocalStackContainer
	if container != nil {
		c = &LocalStackContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// Deprecated: use RunContainer instead
// StartContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func StartContainer(ctx context.Context, overrideReq OverrideContainerRequestOption) (*LocalStackContainer, error) {
	return RunContainer(ctx, overrideReq)
}

func configureDockerHost(req *LocalStackContainerRequest, envVar string) (string, error) {
	reason := ""

	if _, ok := req.Env[envVar]; ok {
		return "explicitly as environment variable", nil
	}

	// if the container is not connected to the default network, use the last network alias in the first network
	// for that we need to check if the container is connected to a network and if it has network aliases
	if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
		alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

		req.Env[envVar] = alias
		return "to match last network alias on container with non-default network", nil
	}

	dockerProvider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return reason, err
	}
	defer dockerProvider.Close()

	daemonHost, err := dockerProvider.DaemonHost(context.Background())
	if err != nil {
		return reason, err
	}

	req.Env[envVar] = daemonHost
	return "to match host-routable address for container", nil
}
