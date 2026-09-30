package kafka

import (
	"crypto/tls"

	"github.com/testcontainers/testcontainers-go"
)

// options holds the configuration settings for the Kafka container.
type options struct {
	tlsEnabled bool
	tlsConfig  *tls.Config
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = Option(nil)

// Option is an option for the Kafka container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithTLS enables an additional SSL listener on the Kafka container.
// When enabled, the container will:
//   - Generate a self-signed CA and a server certificate signed by it,
//     valid for the host the container is reachable at
//   - Configure Kafka with a PKCS12 keystore holding the server certificate
//   - Expose the SSL port (9095), while keeping the PLAINTEXT one
//
// The SSL listener is added to KAFKA_LISTENERS and KAFKA_LISTENER_SECURITY_PROTOCOL_MAP
// after all the options have been applied, so it is kept even if they are overridden.
//
// Use BrokersTLS() to get the SSL endpoint and TLSConfig() to get the
// *tls.Config for client connections.
func WithTLS() Option {
	return func(o *options) error {
		o.tlsEnabled = true
		return nil
	}
}
