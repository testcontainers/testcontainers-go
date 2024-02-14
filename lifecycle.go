package testcontainers

import (
	"context"
	"fmt"
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
// - Readied
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
	PostReadies    []ContainerHook
	PreStops       []ContainerHook
	PostStops      []ContainerHook
	PreTerminates  []ContainerHook
	PostTerminates []ContainerHook
}

// DefaultLoggingHook is a hook that will log the container lifecycle events
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
		PostReadies: []ContainerHook{
			func(ctx context.Context, c Container) error {
				logger.Printf("🔔 Container is ready: %s", shortContainerID(c))
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
				logger.Printf("✅ Container stopped: %s", shortContainerID(c))
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

// defaultPreCreateHook is a hook that will apply the default configuration to the container
var defaultPreCreateHook = func(ctx context.Context, p *DockerProvider, req ContainerRequest, dockerInput *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PreCreates: []ContainerRequestHook{
			func(ctx context.Context, req ContainerRequest) error {
				return p.preCreateContainerHook(ctx, req, dockerInput, hostConfig, networkingConfig)
			},
		},
	}
}

// defaultCopyFileToContainerHook is a hook that will copy files to the container after it's created
// but before it's started
var defaultCopyFileToContainerHook = func(files []ContainerFile) ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PostCreates: []ContainerHook{
			// copy files to container after it's created
			func(ctx context.Context, c Container) error {
				for _, f := range files {
					err := c.CopyFileToContainer(ctx, f.HostFilePath, f.ContainerFilePath, f.FileMode)
					if err != nil {
						return fmt.Errorf("can't copy %s to container: %w", f.HostFilePath, err)
					}
				}

				return nil
			},
		},
	}
}

// defaultLogConsumersHook is a hook that will start log consumers after the container is started
var defaultLogConsumersHook = func(cfg *LogConsumerConfig) ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PostStarts: []ContainerHook{
			// first post-start hook is to produce logs and start log consumers
			func(ctx context.Context, c Container) error {
				dockerContainer := c.(*DockerContainer)

				if cfg == nil {
					return nil
				}

				for _, consumer := range cfg.Consumers {
					dockerContainer.followOutput(consumer)
				}

				if len(cfg.Consumers) > 0 {
					return dockerContainer.startLogProduction(ctx, cfg.Opts...)
				}
				return nil
			},
		},
		PreTerminates: []ContainerHook{
			// first pre-terminate hook is to stop the log production
			func(ctx context.Context, c Container) error {
				if cfg == nil || len(cfg.Consumers) == 0 {
					return nil
				}

				dockerContainer := c.(*DockerContainer)

				return dockerContainer.stopLogProduction()
			},
		},
	}
}

// defaultReadinessHook is a hook that will wait for the container to be ready
var defaultReadinessHook = func() ContainerLifecycleHooks {
	return ContainerLifecycleHooks{
		PostStarts: []ContainerHook{
			// wait for the container to be ready
			func(ctx context.Context, c Container) error {
				dockerContainer := c.(*DockerContainer)

				// if a Wait Strategy has been specified, wait before returning
				if dockerContainer.WaitingFor != nil {
					dockerContainer.logger.Printf(
						"🚧 Waiting for container id %s image: %s. Waiting for: %+v",
						dockerContainer.ID[:12], dockerContainer.Image, dockerContainer.WaitingFor,
					)
					if err := dockerContainer.WaitingFor.WaitUntilReady(ctx, c); err != nil {
						return err
					}
				}

				dockerContainer.isRunning = true

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

// readiedHook is a hook that will be called after a container is ready
func (c *DockerContainer) readiedHook(ctx context.Context) error {
	for _, lifecycleHooks := range c.lifecycleHooks {
		err := containerHookFn(ctx, lifecycleHooks.PostReadies)(c)
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

// Readied is a hook that will be called after a container is ready
func (c ContainerLifecycleHooks) Readied(ctx context.Context) func(container Container) error {
	return containerHookFn(ctx, c.PostReadies)
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
		for p := range image.Config.ExposedPorts {
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

// combineContainerHooks it returns just one ContainerLifecycle hook, as the result of combining
// the default hooks with the user-defined hooks. The function will loop over all the default hooks,
// storing each of the hooks in a slice, and then it will loop over all the user-defined hooks,
// appending or prepending them to the slice of hooks. The order of hooks is the following:
// - for Pre-hooks, always run the default hooks first, then append the user-defined hooks
// - for Post-hooks, always run the user-defined hooks first, then the default hooks
func combineContainerHooks(defaultHooks, userDefinedHooks []ContainerLifecycleHooks) ContainerLifecycleHooks {
	preCreates := []ContainerRequestHook{}
	postCreates := []ContainerHook{}
	preStarts := []ContainerHook{}
	postStarts := []ContainerHook{}
	postReadies := []ContainerHook{}
	preStops := []ContainerHook{}
	postStops := []ContainerHook{}
	preTerminates := []ContainerHook{}
	postTerminates := []ContainerHook{}

	for _, defaultHook := range defaultHooks {
		preCreates = append(preCreates, defaultHook.PreCreates...)
		preStarts = append(preStarts, defaultHook.PreStarts...)
		preStops = append(preStops, defaultHook.PreStops...)
		preTerminates = append(preTerminates, defaultHook.PreTerminates...)
	}

	// append the user-defined hooks after the default pre-hooks
	// and because the post hooks are still empty, the user-defined post-hooks
	// will be the first ones to be executed
	for _, userDefinedHook := range userDefinedHooks {
		preCreates = append(preCreates, userDefinedHook.PreCreates...)
		postCreates = append(postCreates, userDefinedHook.PostCreates...)
		preStarts = append(preStarts, userDefinedHook.PreStarts...)
		postStarts = append(postStarts, userDefinedHook.PostStarts...)
		postReadies = append(postReadies, userDefinedHook.PostReadies...)
		preStops = append(preStops, userDefinedHook.PreStops...)
		postStops = append(postStops, userDefinedHook.PostStops...)
		preTerminates = append(preTerminates, userDefinedHook.PreTerminates...)
		postTerminates = append(postTerminates, userDefinedHook.PostTerminates...)
	}

	// finally, append the default post-hooks
	for _, defaultHook := range defaultHooks {
		postCreates = append(postCreates, defaultHook.PostCreates...)
		postStarts = append(postStarts, defaultHook.PostStarts...)
		postReadies = append(postReadies, defaultHook.PostReadies...)
		postStops = append(postStops, defaultHook.PostStops...)
		postTerminates = append(postTerminates, defaultHook.PostTerminates...)
	}

	return ContainerLifecycleHooks{
		PreCreates:     preCreates,
		PostCreates:    postCreates,
		PreStarts:      preStarts,
		PostStarts:     postStarts,
		PostReadies:    postReadies,
		PreStops:       preStops,
		PostStops:      postStops,
		PreTerminates:  preTerminates,
		PostTerminates: postTerminates,
	}
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
