package cassandra

import (
	"crypto/tls"

	"github.com/testcontainers/testcontainers-go"
)

// options holds the configuration settings for the Cassandra container.
type options struct {
	tlsEnabled bool
	tlsConfig  *tls.Config
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Cassandra container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// defaultOptions returns the default options for the Cassandra container.
func defaultOptions() options {
	return options{
		tlsEnabled: false,
		tlsConfig:  nil,
	}
}

// WithTLS enables TLS/SSL on the Cassandra container.
// When enabled, the container will:
//   - Generate self-signed certificates
//   - Configure Cassandra to use client encryption
//   - Expose the SSL port (9142)
//
// Use TLSConfig() on the returned container to get the *tls.Config for client connections.
func WithTLS() Option {
	return func(o *options) error {
		o.tlsEnabled = true
		return nil
	}
}
