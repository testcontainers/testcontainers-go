package dockermodelrunner

import "github.com/testcontainers/testcontainers-go"

type options struct {
	model string
}

func defaultOptions() options {
	return options{}
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

// WithModel sets the model to pull.
// Multiple calls to this function overrides the previous value.
func WithModel(model string) Option {
	return func(o *options) error {
		o.model = model
		return nil
	}
}
