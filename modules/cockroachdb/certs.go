package cockroachdb

import (
	"crypto/x509"
	"net"
	"time"

	tctls "github.com/testcontainers/testcontainers-go/tls"
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
	caCert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"),
		tctls.WithSubjectCommonName("Cockroach Test CA"),
		tctls.AsCA(),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
	)
	if err != nil {
		return nil, err
	}

	nodeCert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"), // the host will be passed as DNSNames
		tctls.WithSubjectCommonName("node"),
		tctls.AsCA(),
		tctls.WithIPAddresses(net.IPv4(127, 0, 0, 1), net.IPv6loopback),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
		tctls.AsPem(),
		tctls.WithParent(caCert.Cert, caCert.Key),
	)
	if err != nil {
		return nil, err
	}

	clientCert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"),
		tctls.WithSubjectCommonName(defaultUser),
		tctls.AsCA(),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
		tctls.AsPem(),
		tctls.WithParent(caCert.Cert, caCert.Key),
	)
	if err != nil {
		return nil, err
	}

	return &TLSConfig{
		CACert:     caCert.Cert,
		NodeCert:   nodeCert.Bytes,
		NodeKey:    nodeCert.KeyBytes,
		ClientCert: clientCert.Bytes,
		ClientKey:  clientCert.KeyBytes,
	}, nil
}
