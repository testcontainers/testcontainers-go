package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	reuseContainerMx  sync.Mutex
	ErrReuseEmptyName = errors.New("with reuse option a container name mustn't be empty")
)

// GenericContainerRequest represents parameters to a generic container
type GenericContainerRequest struct {
	ContainerRequest              // embedded request for provider
	Started          bool         // whether to auto-start the container
	ProviderType     ProviderType // which provider to use, Docker if empty
	Logger           Logging      // provide a container specific Logging - use default global logger if empty
	Reuse            bool         // reuse an existing container if it exists or create a new one. a container name mustn't be empty
}

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

// WithImage sets the image for a container
func WithImage(image string) CustomizeRequestOption {
	return func(req *GenericContainerRequest) {
		req.Image = image
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

// GenericNetworkRequest represents parameters to a generic network
type GenericNetworkRequest struct {
	NetworkRequest              // embedded request for provider
	ProviderType   ProviderType // which provider to use, Docker if empty
}

// GenericNetwork creates a generic network with parameters
func GenericNetwork(ctx context.Context, req GenericNetworkRequest) (Network, error) {
	provider, err := req.ProviderType.GetProvider()
	if err != nil {
		return nil, err
	}
	network, err := provider.CreateNetwork(ctx, req.NetworkRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create network", err)
	}

	return network, nil
}

// GenericContainer creates a generic container with parameters
func GenericContainer(ctx context.Context, req GenericContainerRequest) (Container, error) {
	if req.Reuse && req.Name == "" {
		return nil, ErrReuseEmptyName
	}

	logging := req.Logger
	if logging == nil {
		logging = Logger
	}
	provider, err := req.ProviderType.GetProvider(WithLogger(logging))
	if err != nil {
		return nil, err
	}

	var c Container
	if req.Reuse {
		// we must protect the reusability of the container in the case it's invoked
		// in a parallel execution, via ParallelContainers or t.Parallel()
		reuseContainerMx.Lock()
		defer reuseContainerMx.Unlock()

		c, err = provider.ReuseOrCreateContainer(ctx, req.ContainerRequest)
	} else {
		c, err = provider.CreateContainer(ctx, req.ContainerRequest)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create container", err)
	}

	if req.Started && !c.IsRunning() {
		if err := c.Start(ctx); err != nil {
			return c, fmt.Errorf("%w: failed to start container", err)
		}
	}
	return c, nil
}

// GenericProvider represents an abstraction for container and network providers
type GenericProvider interface {
	ContainerProvider
	NetworkProvider
}
