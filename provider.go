package testcontainers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

// possible provider types
const (
	ProviderDefault ProviderType = iota // default will auto-detect provider from DOCKER_HOST environment variable
	ProviderDocker
	ProviderPodman // Deprecated: Podman is supported through the current Docker context
)

type (
	// ProviderType is an enum for the possible providers
	ProviderType int

	// GenericProviderOptions defines options applicable to all providers
	GenericProviderOptions struct {
		Logger         Logging
		defaultNetwork string
	}

	// GenericProviderOption defines a common interface to modify GenericProviderOptions
	// These options can be passed to GetProvider in a variadic way to customize the returned GenericProvider instance
	GenericProviderOption interface {
		ApplyGenericTo(opts *GenericProviderOptions)
	}

	// GenericProviderOptionFunc is a shorthand to implement the GenericProviderOption interface
	GenericProviderOptionFunc func(opts *GenericProviderOptions)

	// DockerProviderOptions defines options applicable to DockerProvider
	DockerProviderOptions struct {
		*GenericProviderOptions
	}

	// DockerProviderOption defines a common interface to modify DockerProviderOptions
	// These can be passed to NewDockerProvider in a variadic way to customize the returned DockerProvider instance
	DockerProviderOption interface {
		ApplyDockerTo(opts *DockerProviderOptions)
	}

	// DockerProviderOptionFunc is a shorthand to implement the DockerProviderOption interface
	DockerProviderOptionFunc func(opts *DockerProviderOptions)
)

func (f DockerProviderOptionFunc) ApplyDockerTo(opts *DockerProviderOptions) {
	f(opts)
}

func Generic2DockerOptions(opts ...GenericProviderOption) []DockerProviderOption {
	converted := make([]DockerProviderOption, 0, len(opts))
	for _, o := range opts {
		switch c := o.(type) {
		case DockerProviderOption:
			converted = append(converted, c)
		default:
			converted = append(converted, DockerProviderOptionFunc(func(opts *DockerProviderOptions) {
				o.ApplyGenericTo(opts.GenericProviderOptions)
			}))
		}
	}

	return converted
}

// Deprecated: WithDefaultNetwork is deprecated and will be removed in the next major version
func WithDefaultBridgeNetwork(bridgeNetworkName string) DockerProviderOption {
	return DockerProviderOptionFunc(func(opts *DockerProviderOptions) {
		// NOOP
	})
}

func (f GenericProviderOptionFunc) ApplyGenericTo(opts *GenericProviderOptions) {
	f(opts)
}

// ContainerProvider allows the creation of containers on an arbitrary system
type ContainerProvider interface {
	Close() error                                                                // close the provider
	CreateContainer(context.Context, ContainerRequest) (Container, error)        // create a container without starting it
	ReuseOrCreateContainer(context.Context, ContainerRequest) (Container, error) // reuses a container if it exists or creates a container without starting
	RunContainer(context.Context, ContainerRequest) (Container, error)           // create a container and start it
	Health(context.Context) error
	Config() TestcontainersConfig
}

// GetProvider provides the provider implementation for a certain type
func (t ProviderType) GetProvider(opts ...GenericProviderOption) (GenericProvider, error) {
	opt := &GenericProviderOptions{
		Logger: Logger,
	}

	for _, o := range opts {
		o.ApplyGenericTo(opt)
	}

	provider, err := NewDockerProvider(Generic2DockerOptions(opts...)...)
	if err != nil {
		return nil, fmt.Errorf("%w, failed to create Docker provider", err)
	}
	return provider, nil
}

// NewDockerProvider creates a Docker provider with the EnvClient
func NewDockerProvider(provOpts ...DockerProviderOption) (*DockerProvider, error) {
	o := &DockerProviderOptions{
		GenericProviderOptions: &GenericProviderOptions{
			Logger: Logger,
		},
	}

	for idx := range provOpts {
		provOpts[idx].ApplyDockerTo(o)
	}

	ctx := context.Background()
	c, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &DockerProvider{
		DockerProviderOptions: o,
		host:                  core.MustExtractDockerHost(ctx),
		client:                c,
		config:                cfg,
	}, nil
}
