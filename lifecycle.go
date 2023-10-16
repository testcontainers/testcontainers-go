package testcontainers

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"golang.org/x/exp/slices"
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

var DefaultLoggingHook = func(logger Logging) ContainerLifecycleHooks {
	shortContainerID := func(c Container) string {
		return c.GetContainerID()[:12]
	}

	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, req ContainerRequest) error {
				logger.Printf("🐳 Creating container for image %s", req.Image)
				return nil
			},
		},
		PostCreates: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("✅ Container created: %s", shortContainerID(c))
				return nil
			},
		},
		PreStarts: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("🐳 Starting container: %s", shortContainerID(c))
				return nil
			},
		},
		PostStarts: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("✅ Container started: %s", shortContainerID(c))
				return nil
			},
		},
		PreStops: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("🐳 Stopping container: %s", shortContainerID(c))
				return nil
			},
		},
		PostStops: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("✋ Container stopped: %s", shortContainerID(c))
				return nil
			},
		},
		PreTerminates: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("🐳 Terminating container: %s", shortContainerID(c))
				return nil
			},
		},
		PostTerminates: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("🚫 Container terminated: %s", shortContainerID(c))
				return nil
			},
		},
	}
}

// creatingHook is a hook that will be called before a container is created.
func (req ContainerRequest) creatingHook(ctx context.Context) error {
	for _, lifecycleHooks := range req.LifecycleHooks {
		err := lifecycleHooks.Creating(ctx)(req)
		if err != nil {
			return err
		}
	}

	return nil
}

// createdHook is a hook that will be called after a container is created
func (c *DockerContainer) createdHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PostCreates)(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// startingHook is a hook that will be called before a container is started
func (c *DockerContainer) startingHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PreStarts)(c)
		if err != nil {
			c.printLogs(ctx, err)
			return err
		}
	}

	return nil
}

// startedHook is a hook that will be called after a container is started
func (c *DockerContainer) startedHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PostStarts)(c)
		if err != nil {
			c.printLogs(ctx, err)
			return err
		}
	}

	return nil
}

// printLogs is a helper function that will print the logs of a Docker container
// We are going to use this helper function to inform the user of the logs when an error occurs
func (c *DockerContainer) printLogs(ctx context.Context, cause error) {
	reader, err := c.Logs(ctx)
	if err != nil {
		c.logger.Printf("failed accessing container logs: %v\n", err)
		return
	}

	b, err := io.ReadAll(reader)
	if err != nil {
		c.logger.Printf("failed reading container logs: %v\n", err)
		return
	}

	c.logger.Printf("container logs (%s):\n%s", cause, b)
}

// stoppingHook is a hook that will be called before a container is stopped
func (c *DockerContainer) stoppingHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PreStops)(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// stoppedHook is a hook that will be called after a container is stopped
func (c *DockerContainer) stoppedHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PostStops)(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// terminatingHook is a hook that will be called before a container is terminated
func (c *DockerContainer) terminatingHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PreTerminates)(c)
		if err != nil {
			return err
		}
	}

	return nil
}

// terminatedHook is a hook that will be called after a container is terminated
func (c *DockerContainer) terminatedHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PostTerminates)(c)
		if err != nil {
			return err
		}
	}

	return nil
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

	// only exposing those ports automatically if the container request exposes zero ports and the container does not run in a container network
	if len(exposedPorts) == 0 && !hostConfig.NetworkMode.IsContainer() {
		hostConfig.PortBindings = exposedPortMap
	} else {
		hostConfig.PortBindings = mergePortBindings(hostConfig.PortBindings, exposedPortMap, req.ExposedPorts)
	}

	return nil
}

func mergePortBindings(configPortMap, exposedPortMap nat.PortMap, exposedPorts []string) nat.PortMap {
	if exposedPortMap == nil {
		exposedPortMap = make(map[nat.Port][]nat.PortBinding)
	}

	for k, v := range configPortMap {
		if slices.Contains(exposedPorts, strings.Split(string(k), "/")[0]) {
			exposedPortMap[k] = v
		}
	}
	return exposedPortMap
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
