package cassandra

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/mdelapenya/tlscert"
	"software.sslmate.com/src/go-pkcs12"
)

const (
	// keystorePassword is the default password for the PKCS12 keystore
	keystorePassword = "cassandra"
)

// tlsCerts holds the generated TLS certificates and keystore for Cassandra SSL.
type tlsCerts struct {
	// CACert is the CA certificate
	CACert *tlscert.Certificate
	// ServerCert is the server certificate signed by the CA
	ServerCert *tlscert.Certificate
	// KeystoreBytes is the PKCS12 keystore containing the server certificate and key
	KeystoreBytes []byte
	// TLSConfig is the TLS configuration for Go clients
	TLSConfig *tls.Config
}

// createTLSCerts generates TLS certificates for Cassandra SSL connections.
// It creates:
//   - A self-signed CA certificate
//   - A server certificate signed by the CA
//   - A PKCS12 keystore containing the server cert and key (for Cassandra)
//   - A tls.Config for Go clients to connect securely
func createTLSCerts() (*tlsCerts, error) {
	// IPs to include in the certificates for local testing
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	// Generate CA certificate
	caCert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "Cassandra CA",
		SubjectCommonName: "Cassandra CA",
		IsCA:              true,
	})
	if err != nil {
		return nil, fmt.Errorf("generate CA certificate: %w", err)
	}

	// Generate server certificate signed by CA
	serverCert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "Cassandra Server",
		SubjectCommonName: "localhost",
		Parent:            caCert,
	})
	if err != nil {
		return nil, fmt.Errorf("generate server certificate: %w", err)
	}

	// Create PKCS12 keystore with server cert, key, and CA chain
	// Cassandra 4.0+ supports PKCS12 keystores directly
	// tlscert.Certificate has:
	//   - Cert: *x509.Certificate (parsed certificate)
	//   - Key: *rsa.PrivateKey
	//   - Bytes: []byte (raw certificate bytes)
	keystoreBytes, err := pkcs12.Modern.Encode(
		serverCert.Key,                   // private key
		serverCert.Cert,                  // server certificate
		[]*x509.Certificate{caCert.Cert}, // CA chain
		keystorePassword,
	)
	if err != nil {
		return nil, fmt.Errorf("encode PKCS12 keystore: %w", err)
	}

	// Create TLS config for Go clients
	certPool := x509.NewCertPool()
	certPool.AddCert(caCert.Cert)

	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // Skip hostname verification for container testing
	}

	return &tlsCerts{
		CACert:        caCert,
		ServerCert:    serverCert,
		KeystoreBytes: keystoreBytes,
		TLSConfig:     tlsConfig,
	}, nil
}
