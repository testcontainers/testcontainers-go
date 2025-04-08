package toxiproxy

import (
	"errors"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	portRange int
}

func defaultOptions() options {
	return options{
		portRange: defaultPortRange,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithPortRange sets the port range for the Toxiproxy container.
// Default port range is 31.
func WithPortRange(portRange int) Option {
	return func(o *options) error {
		if portRange < 1 {
			return errors.New("port range must be greater than 0")
		}

		o.portRange = portRange
		return nil
	}
}
