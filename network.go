package testcontainers

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// NetworkProvider allows the creation of networks on an arbitrary system
type NetworkProvider interface {
	CreateNetwork(context.Context, NetworkRequest) (Network, error)            // create a network
	GetNetwork(context.Context, NetworkRequest) (types.NetworkResource, error) // get a network
}

// Network allows getting info about a single network instance
type Network interface {
	Remove(context.Context) error // removes the network
}

type DefaultNetwork string

func (n DefaultNetwork) ApplyGenericTo(opts *GenericProviderOptions) {
	opts.DefaultNetwork = string(n)
}

func (n DefaultNetwork) ApplyDockerTo(opts *DockerProviderOptions) {
	opts.DefaultNetwork = string(n)
}

// NetworkRequest represents the parameters used to get a network
type NetworkRequest struct {
	Driver         string
	CheckDuplicate bool
	Internal       bool
	EnableIPv6     bool
	Name           string
	Labels         map[string]string
	Attachable     bool
	IPAM           *network.IPAM

	SkipReaper    bool              // Deprecated: The reaper is globally controlled by the .testcontainers.properties file or the TESTCONTAINERS_RYUK_DISABLED environment variable
	ReaperImage   string            // Deprecated: use WithImageName ContainerOption instead. Alternative reaper registry
	ReaperOptions []ContainerOption // Deprecated: the reaper is configured at the properties level, for an entire test session
}
