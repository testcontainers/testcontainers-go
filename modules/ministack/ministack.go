package ministack

import (
	"context"
	"fmt"
	"time"

	"github.com/moby/moby/api/types/container"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort         = 4566
	ministackHostEnvVar = "MINISTACK_HOST"
)

// Run creates an instance of the ministack container type
// - overrideReq: a function that can be used to override the default container request, usually used to set the image version, environment variables for ministack, etc.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MinistackContainer, error) {
	dockerHost := testcontainers.MustExtractDockerSocket(ctx)

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 5+len(opts)+1)
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(fmt.Sprintf("%d/tcp", defaultPort)),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/_ministack/health").WithPort("4566/tcp").WithStartupTimeout(120*time.Second)),
		testcontainers.WithEnv(map[string]string{}),
		testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
			hostConfig.Binds = []string{dockerHost + ":/var/run/docker.sock"}
		}),
		testcontainers.WithLogger(log.Default()),
	)

	moduleOpts = append(moduleOpts, opts...)

	// configure the docker host after all the options have been applied
	moduleOpts = append(moduleOpts, configureDockerHost(ctx, ministackHostEnvVar))

	var c *MinistackContainer
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if ctr != nil {
		c = &MinistackContainer{Container: ctr}
	}
	if err != nil {
		return c, fmt.Errorf("run ministack: %w", err)
	}

	return c, nil
}

// configureDockerHost returns an option that configures the docker host environment variable
func configureDockerHost(ctx context.Context, envVar string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		reason, err := setDockerHost(ctx, req, envVar)
		if err != nil {
			return err
		}
		req.Logger.Printf("Setting %s to %s (%s)\n", envVar, req.Env[envVar], reason)
		return nil
	}
}

func setDockerHost(ctx context.Context, req *testcontainers.GenericContainerRequest, envVar string) (string, error) {
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
		return "", err
	}
	defer dockerProvider.Close()

	daemonHost, err := dockerProvider.DaemonHost(ctx)
	if err != nil {
		return "", err
	}

	req.Env[envVar] = daemonHost
	return "to match host-routable address for container", nil
}
