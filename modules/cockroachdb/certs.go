package cockroachdb

import (
	"crypto/rsa"
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
	caCert, caKey, err := generateCA()
	if err != nil {
		return nil, err
	}

	nodeCert, nodeKey, err := generateNode(caCert, caKey)
	if err != nil {
		return nil, err
	}

	clientCert, clientKey, err := generateClient(caCert, caKey)
	if err != nil {
		return nil, err
	}

	return &TLSConfig{
		CACert:     caCert,
		NodeCert:   nodeCert,
		NodeKey:    nodeKey,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}, nil
}

func generateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	caCert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"),
		tctls.WithSubjectCommonName("Cockroach Test CA"),
		tctls.AsCA(),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
	)

	return caCert.Cert, caCert.Key, err
}

func generateNode(caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, []byte, error) {
	cert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"), // the host will be passed as DNSNames
		tctls.WithSubjectCommonName("node"),
		tctls.AsCA(),
		tctls.WithIPAddresses(net.IPv4(127, 0, 0, 1), net.IPv6loopback),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
		tctls.AsPem(),
		tctls.WithParent(caCert, caKey),
	)

	return cert.Bytes, cert.KeyBytes, err
}

func generateClient(caCert *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, []byte, error) {
	cert, err := tctls.GenerateCert(
		tctls.WithHost("localhost"),
		tctls.WithSubjectCommonName(defaultUser),
		tctls.AsCA(),
		tctls.WithValidFrom(time.Now().Add(-time.Hour)),
		tctls.WithValidFor(time.Hour),
		tctls.AsPem(),
		tctls.WithParent(caCert, caKey),
	)

	return cert.Bytes, cert.KeyBytes, err
}
