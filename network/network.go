package network

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/google/uuid"

	"github.com/testcontainers/testcontainers-go/internal/config"
	"github.com/testcontainers/testcontainers-go/internal/core"
	corenetwork "github.com/testcontainers/testcontainers-go/internal/core/network"
	"github.com/testcontainers/testcontainers-go/internal/core/reaper"
)

// New creates a new network with a random UUID name.
// Those existing APIs are deprecated and will be removed in the future, so this function will
// implement the new network APIs when they will be available.
// By default, the network is created with the following options:
// - Driver: bridge
// - Labels: the Testcontainers for Go generic labels, to be managed by Ryuk. Please see the GenericLabels() function
// And those options can be modified by the user, using the CreateModifier function field.
func New(ctx context.Context, opts ...NetworkCustomizer) (*DockerNetwork, error) {
	nc := types.NetworkCreate{
		Driver: "bridge",
		Labels: core.DefaultLabels(core.SessionID()),
	}

	for _, opt := range opts {
		if err := opt.Customize(&nc); err != nil {
			return nil, err
		}
	}

	// Make sure that bridge network exists
	// In case it is disabled we will create reaper_default network
	if _, err := corenetwork.GetDefault(ctx); err != nil {
		return nil, fmt.Errorf("default network not found: %w", err)
	}

	tcConfig := config.Read()

	var err error
	var termSignal chan bool

	if !tcConfig.RyukDisabled {
		termSignal, err = reaper.Connect()
		if err != nil {
			return nil, fmt.Errorf("%w: connecting to network reaper failed", err)
		}
	}

	// Cleanup on error, otherwise set termSignal to nil before successful return.
	defer func() {
		if termSignal != nil {
			termSignal <- true
		}
	}()

	req := corenetwork.Request{
		Driver:     nc.Driver,
		Internal:   nc.Internal,
		EnableIPv6: nc.EnableIPv6,
		Name:       uuid.NewString(),
		Labels:     nc.Labels,
		Attachable: nc.Attachable,
		IPAM:       nc.IPAM,
	}

	response, err := corenetwork.New(ctx, req)
	if err != nil {
		return &DockerNetwork{}, err
	}

	n := &DockerNetwork{
		ID:                response.ID,
		Driver:            req.Driver,
		Name:              req.Name,
		terminationSignal: termSignal,
	}

	// Disable cleanup on success
	termSignal = nil

	return n, nil
}

// NetworkCustomizer is an interface that can be used to configure the network create request.
type NetworkCustomizer interface {
	Customize(req *types.NetworkCreate) error
}

// CustomizeNetworkOption is a type that can be used to configure the network create request.
type CustomizeNetworkOption func(req *types.NetworkCreate) error

// Customize implements the NetworkCustomizer interface,
// applying the option to the network create request.
func (opt CustomizeNetworkOption) Customize(req *types.NetworkCreate) error {
	return opt(req)
}

// WithAttachable allows to set the network as attachable.
func WithAttachable() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		original.Attachable = true

		return nil
	}
}

// WithCheckDuplicate allows to check if a network with the same name already exists.
//
// Deprecated: CheckDuplicate is deprecated since API v1.44, but it defaults to true when sent by the client package to older daemons.
func WithCheckDuplicate() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		return nil
	}
}

// WithDriver allows to override the default network driver, which is "bridge".
func WithDriver(driver string) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		original.Driver = driver

		return nil
	}
}

// WithEnableIPv6 allows to set the network as IPv6 enabled.
// Please use this option if and only if IPv6 is enabled on the Docker daemon.
func WithEnableIPv6() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		original.EnableIPv6 = true

		return nil
	}
}

// WithInternal allows to set the network as internal.
func WithInternal() CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		original.Internal = true

		return nil
	}
}

// WithLabels allows to set the network labels, adding the new ones
// to the default Testcontainers for Go labels.
func WithLabels(labels map[string]string) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		for k, v := range labels {
			original.Labels[k] = v
		}

		return nil
	}
}

// WithIPAM allows to change the default IPAM configuration.
func WithIPAM(ipam *network.IPAM) CustomizeNetworkOption {
	return func(original *types.NetworkCreate) error {
		original.IPAM = ipam

		return nil
	}
}
