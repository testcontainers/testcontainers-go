package container

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

// Create creates a new container, returning the container response from the Docker API.
func Create(
	ctx context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *ocispec.Platform,
	containerName string) (container.CreateResponse, error) {

	var resp container.CreateResponse // zero value
	cli, err := core.NewClient(ctx)
	if err != nil {
		return resp, err
	}
	defer cli.Close()

	resp, err = cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
