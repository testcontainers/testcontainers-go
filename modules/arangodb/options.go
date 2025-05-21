package arangodb

import "github.com/testcontainers/testcontainers-go"

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"ARANGO_ROOT_PASSWORD": defaultPassword,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the ArangoDB container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithRootPassword sets the password for the ArangoDB root user
func WithRootPassword(password string) Option {
	return func(o *options) error {
		o.env["ARANGO_ROOT_PASSWORD"] = password
		return nil
	}
}
