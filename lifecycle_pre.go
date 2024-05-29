package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// defaultPreCreateHook is a hook that will apply the default configuration to the container
var defaultPreCreateHook = func(ctx context.Context, def ContainerDefinition, dockerInput *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, def ContainerDefinition) error {
				return def.PreCreateContainerHook(ctx, dockerInput, hostConfig, networkingConfig)
			},
		},
	}
}
