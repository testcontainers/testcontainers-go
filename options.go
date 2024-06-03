package testcontainers

import (
	"context"
	"fmt"
	"time"

	"dario.cat/mergo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/image"
	"github.com/testcontainers/testcontainers-go/log"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RequestCustomizer is an interface that can be used to configure the Testcontainers container
// request. The passed request will be merged with the default one.
type RequestCustomizer interface {
	Customize(req *Request) error
}

// CustomizeRequestOption is a type that can be used to configure the Testcontainers container request.
// The passed request will be merged with the default one.
type CustomizeRequestOption func(req *Request) error

func (opt CustomizeRequestOption) Customize(req *Request) error {
	return opt(req)
}

// CustomizeRequest returns a function that can be used to merge the passed container request with the one that is used by the container.
// Slices and Maps will be appended.
func CustomizeRequest(src Request) CustomizeRequestOption {
	return func(req *Request) error {
		if err := mergo.Merge(req, &src, mergo.WithOverride, mergo.WithAppendSlice); err != nil {
			return fmt.Errorf("error merging container request, keeping the original one: %w", err)
		}

		return nil
	}
}

// WithConfigModifier allows to override the default container config
func WithConfigModifier(modifier func(config *container.Config)) CustomizeRequestOption {
	return func(req *Request) error {
		req.ConfigModifier = modifier

		return nil
	}
}

// WithEndpointSettingsModifier allows to override the default endpoint settings
func WithEndpointSettingsModifier(modifier func(settings map[string]*network.EndpointSettings)) CustomizeRequestOption {
	return func(req *Request) error {
		req.EnpointSettingsModifier = modifier

		return nil
	}
}

// WithEnv sets the environment variables for a container.
// If the environment variable already exists, it will be overridden.
func WithEnv(envs map[string]string) CustomizeRequestOption {
	return func(req *Request) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}

		for key, val := range envs {
			req.Env[key] = val
		}

		return nil
	}
}

// WithHostConfigModifier allows to override the default host config
func WithHostConfigModifier(modifier func(hostConfig *container.HostConfig)) CustomizeRequestOption {
	return func(req *Request) error {
		req.HostConfigModifier = modifier

		return nil
	}
}

// WithHostPortAccess allows to expose the host ports to the container
func WithHostPortAccess(ports ...int) CustomizeRequestOption {
	return func(req *Request) error {
		if req.HostAccessPorts == nil {
			req.HostAccessPorts = []int{}
		}

		req.HostAccessPorts = append(req.HostAccessPorts, ports...)
		return nil
	}
}

// WithImage sets the image for a container
func WithImage(image string) CustomizeRequestOption {
	return func(req *Request) error {
		req.Image = image

		return nil
	}
}

// WithImageSubstitutors sets the image substitutors for a container
func WithImageSubstitutors(fn ...image.Substitutor) CustomizeRequestOption {
	return func(req *Request) error {
		req.ImageSubstitutors = fn

		return nil
	}
}

// WithLogConsumers sets the log consumers for a container
func WithLogConsumers(consumer ...log.Consumer) CustomizeRequestOption {
	return func(req *Request) error {
		if req.LogConsumerCfg == nil {
			req.LogConsumerCfg = &log.ConsumerConfig{}
		}

		req.LogConsumerCfg.Consumers = consumer
		return nil
	}
}

// Executable represents an executable command to be sent to a container, including options,
// as part of the different lifecycle hooks.
type Executable interface {
	AsCommand() []string
	// Options can container two different types of options:
	// - Docker's ExecConfigs (WithUser, WithWorkingDir, WithEnv, etc.)
	// - testcontainers' ProcessOptions (i.e. Multiplexed response)
	Options() []tcexec.ProcessOption
}

// ExecOptions is a struct that provides a default implementation for the Options method
// of the Executable interface.
type ExecOptions struct {
	opts []tcexec.ProcessOption
}

func (ce ExecOptions) Options() []tcexec.ProcessOption {
	return ce.opts
}

// RawCommand is a type that implements Executable and represents a command to be sent to a container
type RawCommand struct {
	ExecOptions
	cmds []string
}

func NewRawCommand(cmds []string) RawCommand {
	return RawCommand{
		cmds: cmds,
		ExecOptions: ExecOptions{
			opts: []tcexec.ProcessOption{},
		},
	}
}

// AsCommand returns the command as a slice of strings
func (r RawCommand) AsCommand() []string {
	return r.cmds
}

// WithStartupCommand will execute the command representation of each Executable into the container.
// It will leverage the container lifecycle hooks to call the command right after the container
// is started.
func WithStartupCommand(execs ...Executable) CustomizeRequestOption {
	return func(req *Request) error {
		startupCommandsHook := ContainerLifecycleHooks{
			PostStarts: []StartedContainerHook{},
		}

		for _, exec := range execs {
			execFn := func(ctx context.Context, c StartedContainer) error {
				_, _, err := c.Exec(ctx, exec.AsCommand(), exec.Options()...)
				return err
			}

			startupCommandsHook.PostStarts = append(startupCommandsHook.PostStarts, execFn)
		}

		req.LifecycleHooks = append(req.LifecycleHooks, startupCommandsHook)

		return nil
	}
}

// WithAfterReadyCommand will execute the command representation of each Executable into the container.
// It will leverage the container lifecycle hooks to call the command right after the container
// is ready.
func WithAfterReadyCommand(execs ...Executable) CustomizeRequestOption {
	return func(req *Request) error {
		postReadiesHook := []StartedContainerHook{}

		for _, exec := range execs {
			execFn := func(ctx context.Context, c StartedContainer) error {
				_, _, err := c.Exec(ctx, exec.AsCommand(), exec.Options()...)
				return err
			}

			postReadiesHook = append(postReadiesHook, execFn)
		}

		req.LifecycleHooks = append(req.LifecycleHooks, ContainerLifecycleHooks{
			PostReadies: postReadiesHook,
		})

		return nil
	}
}

// WithWaitStrategy sets the wait strategy for a container, using 60 seconds as deadline
func WithWaitStrategy(strategies ...wait.Strategy) CustomizeRequestOption {
	return WithWaitStrategyAndDeadline(60*time.Second, strategies...)
}

// WithWaitStrategyAndDeadline sets the wait strategy for a container, including deadline
func WithWaitStrategyAndDeadline(deadline time.Duration, strategies ...wait.Strategy) CustomizeRequestOption {
	return func(req *Request) error {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(deadline)

		return nil
	}
}

// WithLogger sets the logger to be used for the container
func WithLogger(logger log.Logging) CustomizeRequestOption {
	return func(req *Request) error {
		req.Logger = logger

		return nil
	}
}

// ----------------------------------------------------------------------------
// Network Options
// ----------------------------------------------------------------------------

// WithNetwork reuses an already existing network, attaching the container to it.
// Finally it sets the network alias on that network to the given alias.
func WithNetwork(aliases []string, nw *DockerNetwork) CustomizeRequestOption {
	return func(req *Request) error {
		networkName := nw.Name

		// attaching to the network because it was created with success or it already existed.
		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = aliases

		return nil
	}
}

// WithNewNetwork creates a new network with random name and customizers, and attaches the container to it.
// Finally it sets the network alias on that network to the given alias.
func WithNewNetwork(ctx context.Context, aliases []string, opts ...tcnetwork.Customizer) CustomizeRequestOption {
	return func(req *Request) error {
		newNetwork, err := NewNetwork(ctx, opts...)
		if err != nil {
			return fmt.Errorf("new network: %w", err)
		}

		networkName := newNetwork.Name

		// attaching to the network because it was created with success or it already existed.
		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = aliases

		return nil
	}
}
