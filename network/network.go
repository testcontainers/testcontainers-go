package network

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go"
)

// New creates a new network with a random UUID name, calling the already existing GenericNetwork APIs.
// Those existing APIs are deprecated and will be removed in the future, so this function will
// implement the new network APIs when they will be available.
// By default, the network is created with the following options:
// - Driver: bridge
// - Labels: the Testcontainers for Go generic labels, to be managed by Ryuk. Please see the GenericLabels() function
// And those options can be modified by the user, using the CreateModifier function field.
func New(ctx context.Context, opts ...NetworkCustomizer) (*testcontainers.DockerNetwork, error) {
	nc := types.NetworkCreate{
		Driver: "bridge",
		Labels: testcontainers.GenericLabels(),
	}

	for _, opt := range opts {
		opt.Customize(&nc)
	}

	//nolint:staticcheck
	netReq := testcontainers.NetworkRequest{
		Driver:         nc.Driver,
		CheckDuplicate: nc.CheckDuplicate,
		Internal:       nc.Internal,
		EnableIPv6:     nc.EnableIPv6,
		Name:           uuid.NewString(),
		Labels:         nc.Labels,
		Attachable:     nc.Attachable,
		IPAM:           nc.IPAM,
	}

	//nolint:staticcheck
	n, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: netReq,
	})
	if err != nil {
		return nil, err
	}

	// Return a DockerNetwork struct instead of the Network interface,
	// following the "accept interface, return struct" pattern.
	return n.(*testcontainers.DockerNetwork), nil
}

// NetworkCustomizer is an interface that can be used to configure the network create request.
type NetworkCustomizer interface {
	Customize(req *types.NetworkCreate)
}

// CustomizeNetworkOption is a type that can be used to configure the network create request.
type CustomizeNetworkOption func(req *types.NetworkCreate)

// Customize implements the NetworkCustomizer interface,
// applying the option to the network create request.
func (opt CustomizeNetworkOption) Customize(req *types.NetworkCreate) {
	opt(req)
}

// WithAttachable allows to set the network as attachable.
func WithAttachable() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		original.Attachable = true
	}
}

// WithCheckDuplicate allows to check if a network with the same name already exists.
func WithCheckDuplicate() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		//nolint:staticcheck
		original.CheckDuplicate = true
	}
}

// WithDriver allows to override the default network driver, which is "bridge".
func WithDriver(driver string) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		original.Driver = driver
	}
}

// WithEnableIPv6 allows to set the network as IPv6 enabled.
// Please use this option if and only if IPv6 is enabled on the Docker daemon.
func WithEnableIPv6() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		original.EnableIPv6 = true
	}
}

// WithInternal allows to set the network as internal.
func WithInternal() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		original.Internal = true
	}
}

// WithLabels allows to set the network labels, adding the new ones
// to the default Testcontainers for Go labels.
func WithLabels(labels map[string]string) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		for k, v := range labels {
			original.Labels[k] = v
		}
	}
}

// WithIPAM allows to change the default IPAM configuration.
func WithIPAM(ipam *network.IPAM) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) {
		original.IPAM = ipam
	}
}

// WithNetwork reuses an already existing network, attaching the container to it.
// Finally it sets the network alias on that network to the given alias.
func WithNetwork(aliases []string, nw *testcontainers.DockerNetwork) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		networkName := nw.Name

		// attaching to the network because it was created with success or it already existed.
		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = aliases
	}
}

// WithNewNetwork creates a new network with random name and customizers, and attaches the container to it.
// Finally it sets the network alias on that network to the given alias.
func WithNewNetwork(ctx context.Context, aliases []string, opts ...NetworkCustomizer) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		newNetwork, err := New(ctx, opts...)
		if err != nil {
			logger := req.Logger
			if logger == nil {
				logger = testcontainers.Logger
			}
			logger.Printf("failed to create network. Container won't be attached to it: %v", err)
			return
		}

		networkName := newNetwork.Name

		// attaching to the network because it was created with success or it already existed.
		req.Networks = append(req.Networks, networkName)

		if req.NetworkAliases == nil {
			req.NetworkAliases = make(map[string][]string)
		}
		req.NetworkAliases[networkName] = aliases
	}
}
