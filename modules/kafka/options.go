package kafka

import (
	"github.com/testcontainers/testcontainers-go"
)

// options holds the configuration settings for the Kafka container.
type options struct {
	tlsEnabled bool
	// tlsCerts is generated on the first start and reused on restarts,
	// so the keystore in the container always matches TLSConfig()
	tlsCerts *tlsCerts
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
// If they already define an SSL listener, it must listen on port 9095 and use the
// SSL protocol, otherwise the container fails to start with an error.
//
// Use BrokersTLS() to get the SSL endpoint and TLSConfig() to get the
// *tls.Config for client connections.
func WithTLS() Option {
	return func(o *options) error {
		o.tlsEnabled = true
		return nil
	}
}
