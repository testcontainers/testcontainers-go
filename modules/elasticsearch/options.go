package elasticsearch

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// EnableTLS is a flag to enable TLS.
	EnableTLS bool
	certBytes []byte
	Password  string
}

func defaultOptions() options {
	return options{
		EnableTLS: false,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithPassword sets the password for the Elasticsearch container.
func WithPassword(password string) Option {
	return func(o *options) {
		o.Password = password
	}
}
