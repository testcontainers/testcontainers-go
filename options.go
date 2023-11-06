package testcontainers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerCustomizer is an interface that can be used to configure the Testcontainers container
// request. The passed request will be merged with the default one.
type ContainerCustomizer interface {
	Customize(req *GenericContainerRequest)
}

// CustomizeRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
type CustomizeRequestOption func(req *GenericContainerRequest)

func (opt CustomizeRequestOption) Customize(req *GenericContainerRequest) {
	opt(req)
}

// CustomizeRequest returns a function that can be used to merge the passed container request with the one that is used by the container.
// Slices and Maps will be appended.
func CustomizeRequest(src GenericContainerRequest) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		if err := mergo.Merge(req, &src, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			fmt.Printf("error merging container request, keeping the original one. Error: %v", err)
			return
		}
	}
}

// WithConfigModifier allows to override the default container config
func WithConfigModifier(modifier func(config *container.Config)) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.ConfigModifier = modifier
	}
}

// WithEndpointSettingsModifier allows to override the default endpoint settings
func WithEndpointSettingsModifier(modifier func(settings map[string]*network.EndpointSettings)) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.EnpointSettingsModifier = modifier
	}
}

// WithHostConfigModifier allows to override the default host config
func WithHostConfigModifier(modifier func(hostConfig *container.HostConfig)) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.HostConfigModifier = modifier
	}
}

// WithImage sets the image for a container
func WithImage(image string) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.Image = image
	}
}

// imageSubstitutor {
// ImageSubstitutor represents a way to substitute container image names
type ImageSubstitutor interface {
	// Description returns the name of the type and a short description of how it modifies the image.
	// Useful to be printed in logs
	Description() string
	Substitute(image string) (string, error)
}

// }

// WithImageSubstitutors sets the image substitutors for a container
func WithImageSubstitutors(fn ...ImageSubstitutor) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.ImageSubstitutors = fn
	}
}

// WithNetwork creates a network with the given name and attaches the container to it, setting the network alias
// on that network to the given alias.
// If the network already exists, checking if the network name already exists, it will be reused.
func WithNetwork(networkName string, alias string) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		_, err := GenericNetwork(context.Background(), GenericNetworkRequest{
			NetworkRequest: NetworkRequest{
				Name:           networkName,
				CheckDuplicate: true, // force the Docker provider to reuse an existing network
			},
		})
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			logger := req.Logger
			if logger == nil {
				logger = Logger
			}
			logger.Printf("Failed to create network '%s'. Container won't be attached to this network: %v", networkName, err)
			return
		}

		// attaching to the network because it was created with success or it already existed.
		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = []string{alias}
	}
}

// Executable represents an executable command to be sent to a container
// as part of the PostStart lifecycle hook.
type Executable interface {
	AsCommand() []string
}

// RawCommand is a type that implements Executable and represents a command to be sent to a container
type RawCommand []string

// AsCommand returns the command as a slice of strings
func (r RawCommand) AsCommand() []string {
	return r
}

// WithStartupCommand will execute the command representation of each Executable into the container.
// It will leverage the container lifecycle hooks to call the command right after the container
// is started.
func WithStartupCommand(execs ...Executable) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		startupCommandsHook := ContainerLifecycleHooks{
			PostStarts: []ContainerHook{},
		}

		for _, exec := range execs {
			execFn := func(ctx context.Context, c Container) error {
				_, _, err := c.Exec(ctx, exec.AsCommand())
				return err
			}

			startupCommandsHook.PostStarts = append(startupCommandsHook.PostStarts, execFn)
		}

		req.LifecycleHooks = append(req.LifecycleHooks, startupCommandsHook)
	}
}

// WithWaitStrategy sets the wait strategy for a container, using 60 seconds as deadline
func WithWaitStrategy(strategies ...wait.Strategy) CustomizeRequestOption {
	return WithWaitStrategyAndDeadline(60*time.Second, strategies...)
}

// WithWaitStrategyAndDeadline sets the wait strategy for a container, including deadline
func WithWaitStrategyAndDeadline(deadline time.Duration, strategies ...wait.Strategy) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(deadline)
	}
}
