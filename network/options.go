package network

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// Customizer is an interface that can be used to configure the network create request.
type Customizer interface {
	Customize(req *types.NetworkCreate) error
}

// CustomizeOption is a type that can be used to configure the network create request.
type CustomizeOption func(req *types.NetworkCreate) error

// Customize implements the NetworkCustomizer interface,
// applying the option to the network create request.
func (opt CustomizeOption) Customize(req *types.NetworkCreate) error {
	return opt(req)
}

// WithAttachable allows to set the network as attachable.
func WithAttachable() CustomizeOption {
	return func(original *types.NetworkCreate) error {
		original.Attachable = true

		return nil
	}
}

// WithDriver allows to override the default network driver, which is "bridge".
func WithDriver(driver string) CustomizeOption {
	return func(original *types.NetworkCreate) error {
		original.Driver = driver

		return nil
	}
}

// WithEnableIPv6 allows to set the network as IPv6 enabled.
// Please use this option if and only if IPv6 is enabled on the Docker daemon.
func WithEnableIPv6() CustomizeOption {
	return func(original *types.NetworkCreate) error {
		original.EnableIPv6 = true

		return nil
	}
}

// WithInternal allows to set the network as internal.
func WithInternal() CustomizeOption {
	return func(original *types.NetworkCreate) error {
		original.Internal = true

		return nil
	}
}

// WithLabels allows to set the network labels, adding the new ones
// to the default Testcontainers for Go labels.
func WithLabels(labels map[string]string) CustomizeOption {
	return func(original *types.NetworkCreate) error {
		for k, v := range labels {
			original.Labels[k] = v
		}

		return nil
	}
}

// WithIPAM allows to change the default IPAM configuration.
func WithIPAM(ipam *network.IPAM) CustomizeOption {
	return func(original *types.NetworkCreate) error {
		original.IPAM = ipam

		return nil
	}
}
