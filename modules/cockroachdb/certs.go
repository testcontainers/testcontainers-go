package cockroachdb

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/mdelapenya/tlscert"

	"github.com/testcontainers/testcontainers-go"
)

// TLSConfig is a [testcontainers.ContainerCustomizer] that enables TLS for CockroachDB.
type TLSConfig struct {
	CACert     *x509.Certificate
	NodeCert   []byte
	NodeKey    []byte
	ClientCert []byte
	ClientKey  []byte

	cfg *tls.Config
}

// customize implements the [customizer] interface.
// It sets the TLS config on the CockroachDBContainer.
func (c *TLSConfig) customize(ctr *CockroachDBContainer) error {
	ctr.tlsConfig = c.cfg
	return nil
}

// Customize implements the [testcontainers.ContainerCustomizer] interface.
func (c *TLSConfig) Customize(req *testcontainers.GenericContainerRequest) error {
	if req.Env[envUser] != defaultUser {
		return fmt.Errorf("unsupported user %q with TLS, use %q", req.Env[envUser], defaultUser)
	}

	req.Env[envOptionTLS] = "true"

	caBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.CACert.Raw,
	})
	files := map[string][]byte{
		fileCACert:     caBytes,
		fileNodeCert:   c.NodeCert,
		fileNodeKey:    c.NodeKey,
		fileClientCert: c.ClientCert,
		fileClientKey:  c.ClientKey,
	}

	for filename, contents := range files {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            bytes.NewReader(contents),
			ContainerFilePath: filename,
			FileMode:          0o600,
		})
	}

	req.Cmd = append(req.Cmd, "--certs-dir="+certsDir)

	return nil
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
		return nil, errors.New("failed to generate CA certificate")
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
		return nil, errors.New("failed to generate node certificate")
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
		return nil, errors.New("failed to generate client certificate")
	}

	keyPair, err := tls.X509KeyPair(clientCert.Bytes, clientCert.KeyBytes)
	if err != nil {
		return nil, fmt.Errorf("x509 key pair: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert.Cert)

	return &TLSConfig{
		CACert:     caCert.Cert,
		NodeCert:   nodeCert.Bytes,
		NodeKey:    nodeCert.KeyBytes,
		ClientCert: clientCert.Bytes,
		ClientKey:  clientCert.KeyBytes,
		cfg: &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{keyPair},
			ServerName:   "localhost",
		},
	}, nil
}
