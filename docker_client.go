package testcontainers

import (
	"context"

	"github.com/docker/docker/client"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// DockerClient is a wrapper around the docker client that is used by testcontainers-go.
// It implements the SystemAPIClient interface in order to cache the docker info and reuse it.
type DockerClient = core.DockerClient

func NewDockerClientWithOpts(ctx context.Context, opt ...client.Opt) (*DockerClient, error) {
	dockerClient, err := core.NewClient(ctx, opt...)
	if err != nil {
		return nil, err
	}

	return dockerClient, nil
}
