package testcontainers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"dario.cat/mergo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
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

// prependHubRegistry represents a way to prepend a custom Hub registry to the image name,
// using the HubImageNamePrefix configuration value
type prependHubRegistry struct {
	prefix string
}

// newPrependHubRegistry creates a new prependHubRegistry
func newPrependHubRegistry() prependHubRegistry {
	hubPrefix := config.Read().HubImageNamePrefix

	return prependHubRegistry{
		prefix: hubPrefix,
	}
}

// Description returns the name of the type and a short description of how it modifies the image.
func (p prependHubRegistry) Description() string {
	return fmt.Sprintf("HubImageSubstitutor (prepends %s)", p.prefix)
}

// Substitute prepends the Hub prefix to the image name, with certain conditions:
//   - if the prefix is empty, the image is returned as is.
//   - if the image is a non-hub image (e.g. where another registry is set), the image is returned as is.
//   - if the image is a Docker Hub image where the hub registry is explicitly part of the name
//     (i.e. anything with a docker.io or registry.hub.docker.com host part), the image is returned as is.
func (p prependHubRegistry) Substitute(image string) (string, error) {
	registry := testcontainersdocker.ExtractRegistry(image, "")

	// add the exclusions in the right order
	exclusions := []func() bool{
		func() bool { return p.prefix == "" },                        // no prefix set at the configuration level
		func() bool { return registry != "" },                        // non-hub image
		func() bool { return registry == "docker.io" },               // explicitly including docker.io
		func() bool { return registry == "registry.hub.docker.com" }, // explicitly including registry.hub.docker.com
	}

	for _, exclusion := range exclusions {
		if exclusion() {
			return image, nil
		}
	}

	return fmt.Sprintf("%s/%s", p.prefix, image), nil
}

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
	return func(req *GenericContainerRequest) {
		startupCommandsHook := ContainerLifecycleHooks{
			PostStarts: []ContainerHook{},
		}

		for _, exec := range execs {
			execFn := func(ctx context.Context, c Container) error {
				_, _, err := c.Exec(ctx, exec.AsCommand(), exec.Options()...)
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
