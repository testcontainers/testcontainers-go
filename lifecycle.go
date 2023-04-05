package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// ContainerRequestHook is a hook that will be called before a container is created.
// It can be used to modify container configuration before it is created,
// using the different lifecycle hooks that are available:
// - Creating
// For that, it will receive a ContainerRequest, modify it and return an error if needed.
type ContainerRequestHook func(ctx context.Context, req ContainerRequest) error

// ContainerHook is a hook that will be called after a container is created
// It can be used to modify the state of the container after it is created,
// using the different lifecycle hooks that are available:
// - Created
// - Starting
// - Started
// - Stopping
// - Stopped
// - Terminating
// - Terminated
// For that, it will receive a Container, modify it and return an error if needed.
type ContainerHook func(ctx context.Context, container Container) error

// ContainerLifecycleHooks is a struct that contains all the hooks that can be used
// to modify the container lifecycle. All the container lifecycle hooks except the PreCreates hooks
// will be passed to the container once it's created
type ContainerLifecycleHooks struct {
	PreCreates     []ContainerRequestHook
	PostCreates    []ContainerHook
	PreStarts      []ContainerHook
	PostStarts     []ContainerHook
	PreStops       []ContainerHook
	PostStops      []ContainerHook
	PreTerminates  []ContainerHook
	PostTerminates []ContainerHook
}

// Creating is a hook that will be called before a container is created.
func (c ContainerLifecycleHooks) Creating(ctx context.Context) func(req ContainerRequest) error {
	return func(req ContainerRequest) error {
		for _, hook := range c.PreCreates {
			if err := hook(ctx, req); err != nil {
				return err
			}
		}

		return nil
	}
}

// containerHookFn is a helper function that will create a function to be returned by all the different
// container lifecycle hooks. The created function will iterate over all the hooks and call them one by one.
func containerHookFn(ctx context.Context, containerHook []ContainerHook) func(container Container) error {
	return func(container Container) error {
		for _, hook := range containerHook {
			if err := hook(ctx, container); err != nil {
				return err
			}
		}

		return nil
	}
}

// Created is a hook that will be called after a container is created
func (c ContainerLifecycleHooks) Created(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PostCreates)
}

// Starting is a hook that will be called before a container is started
func (c ContainerLifecycleHooks) Starting(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PreStarts)
}

// Started is a hook that will be called after a container is started
func (c ContainerLifecycleHooks) Started(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PostStarts)
}

// Stopping is a hook that will be called before a container is stopped
func (c ContainerLifecycleHooks) Stopping(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PreStops)
}

// Stopped is a hook that will be called after a container is stopped
func (c ContainerLifecycleHooks) Stopped(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PostStops)
}

// Terminating is a hook that will be called before a container is terminated
func (c ContainerLifecycleHooks) Terminating(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PreTerminates)
}

// Terminated is a hook that will be called after a container is terminated
func (c ContainerLifecycleHooks) Terminated(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PostTerminates)
}

func (p *DockerProvider) preCreateContainerHook(ctx context.Context, req ContainerRequest, dockerInput *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) error {
	// prepare mounts
	hostConfig.Mounts = mapToDockerMounts(req.Mounts)

	endpointSettings := map[string]*network.EndpointSettings{}

	// #248: Docker allows only one network to be specified during container creation
	// If there is more than one network specified in the request container should be attached to them
	// once it is created. We will take a first network if any specified in the request and use it to create container
	if len(req.Networks) > 0 {
		attachContainerTo := req.Networks[0]

		nw, err := p.GetNetwork(ctx, NetworkRequest{
			Name: attachContainerTo,
		})
		if err == nil {
			aliases := []string{}
			if _, ok := req.NetworkAliases[attachContainerTo]; ok {
				aliases = req.NetworkAliases[attachContainerTo]
			}
			endpointSetting := network.EndpointSettings{
				Aliases:   aliases,
				NetworkID: nw.ID,
			}
			endpointSettings[attachContainerTo] = &endpointSetting
		}
	}

	if req.ConfigModifier != nil {
		req.ConfigModifier(dockerInput)
	}

	if req.HostConfigModifier == nil {
		req.HostConfigModifier = defaultHostConfigModifier(req)
	}
	req.HostConfigModifier(hostConfig)

	if req.EnpointSettingsModifier != nil {
		req.EnpointSettingsModifier(endpointSettings)
	}

	networkingConfig.EndpointsConfig = endpointSettings

	exposedPorts := req.ExposedPorts
	// this check must be done after the pre-creation Modifiers are called, so the network mode is already set
	if len(exposedPorts) == 0 && !hostConfig.NetworkMode.IsContainer() {
		image, _, err := p.client.ImageInspectWithRaw(ctx, dockerInput.Image)
		if err != nil {
			return err
		}
		for p := range image.ContainerConfig.ExposedPorts {
			exposedPorts = append(exposedPorts, string(p))
		}
	}

	exposedPortSet, exposedPortMap, err := nat.ParsePortSpecs(exposedPorts)
	if err != nil {
		return err
	}

	dockerInput.ExposedPorts = exposedPortSet
	hostConfig.PortBindings = exposedPortMap

	return nil
}

// defaultHostConfigModifier provides a default modifier including the deprecated fields
func defaultHostConfigModifier(req ContainerRequest) func(hostConfig *container.HostConfig) {
	return func(hostConfig *container.HostConfig) {
		hostConfig.AutoRemove = req.AutoRemove
		hostConfig.CapAdd = req.CapAdd
		hostConfig.CapDrop = req.CapDrop
		hostConfig.Binds = req.Binds
		hostConfig.ExtraHosts = req.ExtraHosts
		hostConfig.NetworkMode = req.NetworkMode
		hostConfig.Resources = req.Resources
	}
}
