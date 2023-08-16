package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

type TestcontainersClient struct {
	*client.Client // client is embedded into our own client
}

// implements SystemAPIClient interface
var _ client.SystemAPIClient = &TestcontainersClient{}

// Events returns a channel to listen to events that happen to the docker daemon.
func (c *TestcontainersClient) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	return c.Client.Events(ctx, options)
}

// Info returns information about the docker server.
func (c *TestcontainersClient) Info(ctx context.Context) (types.Info, error) {
	return c.Client.Info(ctx)
}

// RegistryLogin logs into a Docker registry.
func (c *TestcontainersClient) RegistryLogin(ctx context.Context, auth registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	return c.Client.RegistryLogin(ctx, auth)
}

// DiskUsage returns the disk usage of all images.
func (c *TestcontainersClient) DiskUsage(ctx context.Context, options types.DiskUsageOptions) (types.DiskUsage, error) {
	return c.Client.DiskUsage(ctx, options)
}

// Ping pings the docker server.
func (c *TestcontainersClient) Ping(ctx context.Context) (types.Ping, error) {
	return c.Client.Ping(ctx)
}

// Deprecated: Use NewTestcontainersClient instead.
func NewDockerClient() (*client.Client, error) {
	cli, err := NewTestcontainersClient(context.Background())
	if err != nil {
		return nil, err
	}

	return cli.Client, nil
}

func NewTestcontainersClient(ctx context.Context, opt ...client.Opt) (*TestcontainersClient, error) {
	dockerClient, err := testcontainersdocker.NewClient(ctx, opt...)
	if err != nil {
		return nil, err
	}

	tcClient := TestcontainersClient{
		Client: dockerClient,
	}

	if _, err = tcClient.Info(context.Background()); err != nil {
		// Fallback to environment, including the original options
		if len(opt) == 0 {
			opt = []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
		}

		dockerClient, err := client.NewClientWithOpts(opt...)
		if err != nil {
			return nil, err
		}

		tcClient.Client = dockerClient
	}
	defer tcClient.Close()

	return &tcClient, nil
}
