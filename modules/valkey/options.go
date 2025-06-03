package valkey

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/mdelapenya/tlscert"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	tlsEnabled bool
	tlsConfig  *tls.Config
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

// WithTLS sets the TLS configuration for the redis container, setting
// the 6380/tcp port to listen on for TLS connections and using a secure URL (rediss://).
func WithTLS() Option {
	return func(o *options) error {
		o.tlsEnabled = true
		return nil
	}
}

// createTLSCerts creates a CA certificate, a client certificate and a Valkey certificate.
func createTLSCerts() (caCert *tlscert.Certificate, clientCert *tlscert.Certificate, valkeyCert *tlscert.Certificate, err error) {
	// ips is the extra list of IPs to include in the certificates.
	// It's used to allow the client and Valkey certificates to be used in the same host
	// when the tests are run using a remote docker daemon.
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	// Generate CA certificate
	caCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "ca",
		SubjectCommonName: "ca",
		IsCA:              true,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate CA certificate: %w", err)
	}

	// Generate client certificate
	clientCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		Name:              "Redis Client",
		SubjectCommonName: "localhost",
		IPAddresses:       ips,
		Parent:            caCert,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate client certificate: %w", err)
	}

	// Generate Valkey certificate
	valkeyCert, err = tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:        "localhost",
		IPAddresses: ips,
		Name:        "Valkey Server",
		Parent:      caCert,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate Valkey certificate: %w", err)
	}

	return caCert, clientCert, valkeyCert, nil
}
