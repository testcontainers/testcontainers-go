package localstack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/mod/semver"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort            = 4566
	defaultVersion         = "1.4.0"
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
func WithNetwork(networkName string, alias string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		_, err := testcontainers.GenericNetwork(context.Background(), testcontainers.GenericNetworkRequest{
			NetworkRequest: testcontainers.NetworkRequest{
				Name: networkName,
			},
		})
		if err != nil {
			logger := req.Logger
			if logger == nil {
				logger = testcontainers.Logger
			}
			logger.Printf("Failed to create network '%s'. Container won't be attached to this network: %v", networkName, err)
			return
		}

		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = []string{alias}
	}
}

// RunContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*LocalStackContainer, error) {
	dockerHost := testcontainersdocker.ExtractDockerSocket(ctx)

	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("localstack/localstack:%s", defaultVersion),
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
		opt.Customize(&localStackReq.GenericContainerRequest)
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
	if err != nil {
		return nil, err
	}

	c := &LocalStackContainer{
		Container: container,
	}
	return c, nil
}

// StartContainer creates an instance of the LocalStack container type, being possible to pass a custom request and options:
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for localstack, etc.
// Deprecated: use RunContainer instead
func StartContainer(ctx context.Context, overrideReq OverrideContainerRequestOption) (*LocalStackContainer, error) {
	return RunContainer(ctx, overrideReq)
}

func configureDockerHost(req *LocalStackContainerRequest, envVar string) (reason string, err error) {
	err = nil
	reason = ""

	if _, ok := req.Env[envVar]; ok {
		reason = "explicitly as environment variable"
		return
	}

	// if the container is not connected to the default network, use the last network alias in the first network
	// for that we need to check if the container is connected to a network and if it has network aliases
	if len(req.Networks) > 0 && len(req.NetworkAliases) > 0 && len(req.NetworkAliases[req.Networks[0]]) > 0 {
		alias := req.NetworkAliases[req.Networks[0]][len(req.NetworkAliases[req.Networks[0]])-1]

		req.Env[envVar] = alias
		reason = "to match last network alias on container with non-default network"
		return
	}

	var dockerProvider *testcontainers.DockerProvider
	dockerProvider, err = testcontainers.NewDockerProvider()
	if err != nil {
		return
	}
	defer dockerProvider.Close()

	var daemonHost string
	daemonHost, err = dockerProvider.DaemonHost(context.Background())
	if err != nil {
		return
	}

	req.Env[envVar] = daemonHost
	reason = "to match host-routable address for container"

	return
}
