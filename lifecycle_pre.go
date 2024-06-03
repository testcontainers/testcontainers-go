package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

// defaultPreCreateHook is a hook that will apply the default configuration to the container
var defaultPreCreateHook = func(dockerInput *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) LifecycleHooks {
	return LifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, req *Request) error {
				return req.preCreateContainerHook(ctx, dockerInput, hostConfig, networkingConfig)
			},
		},
	}
}
