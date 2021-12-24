package testcontainers

import (
	"context"
	"fmt"
)

// GenericContainerRequest represents parameters to a generic container
type GenericContainerRequest struct {
	ContainerRequest              // embedded request for provider
	Started          bool         // whether to auto-start the container
	ProviderType     ProviderType // which provider to use, Docker if empty
	Logger           Logging      // provide a container specific Logging - use default global logger if empty
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
	logging := req.Logger
	if logging == nil {
		logging = Logger
	}
	provider, err := req.ProviderType.GetProvider(WithLogger(logging))
	if err != nil {
		return nil, err
	}

	c, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create container", err)
	}

	if req.Started {
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
