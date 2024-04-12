package cockroachdb

import (
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/mdelapenya/tlscert"
)

type TLSConfig struct {
	CACert     *x509.Certificate
	NodeCert   []byte
	NodeKey    []byte
	ClientCert []byte
	ClientKey  []byte
}

// NewTLSConfig creates a new TLSConfig capable of running CockroachDB & connecting over TLS.
func NewTLSConfig() (*TLSConfig, error) {
	// exampleSelfSignedCert {
	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:              "ca",
		SubjectCommonName: "Cockroach Test CA",
		Host:              "localhost,127.0.0.1",
		IsCA:              true,
		ValidFor:          time.Hour,
	})
	if caCert == nil {
		return nil, fmt.Errorf("failed to generate CA certificate")
	}
	// }

	// exampleSignSelfSignedCert {
	nodeCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:              "node",
		SubjectCommonName: "node",
		Host:              "localhost,127.0.0.1",
		IPAddresses:       []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		ValidFor:          time.Hour,
		Parent:            caCert, // using the CA certificate as parent
	})
	if nodeCert == nil {
		return nil, fmt.Errorf("failed to generate node certificate")
	}
	// }

	clientCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:              "client",
		SubjectCommonName: defaultUser,
		Host:              "localhost,127.0.0.1",
		ValidFor:          time.Hour,
		Parent:            caCert, // using the CA certificate as parent
	})
	if clientCert == nil {
		return nil, fmt.Errorf("failed to generate client certificate")
	}

	return &TLSConfig{
		CACert:     caCert.Cert,
		NodeCert:   nodeCert.Bytes,
		NodeKey:    nodeCert.KeyBytes,
		ClientCert: clientCert.Bytes,
		ClientKey:  clientCert.KeyBytes,
	}, nil
}
