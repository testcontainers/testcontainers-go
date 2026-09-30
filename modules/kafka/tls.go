package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/mdelapenya/tlscert"
	"software.sslmate.com/src/go-pkcs12"
)

const (
	// keystorePassword is the password for the PKCS12 keystore and the key inside it.
	keystorePassword = "testcontainers"

	// secretsDir is the directory where the Confluent image expects the SSL material.
	secretsDir = "/etc/kafka/secrets"

	keystoreFilename    = "kafka.keystore.p12"
	credentialsFilename = "keystore_creds"
)

// tlsCerts holds the generated TLS material for the Kafka SSL listener.
type tlsCerts struct {
	// KeystoreBytes is the PKCS12 keystore containing the server certificate, key and CA chain
	KeystoreBytes []byte
	// TLSConfig is the TLS configuration for Go clients
	TLSConfig *tls.Config
}

// createTLSCerts generates a CA and a server certificate signed by it.
// The server certificate is valid for localhost, 127.0.0.1 and the given host,
// which is the host the container is reachable at (e.g. a remote docker host).
func createTLSCerts(host string) (*tlsCerts, error) {
	hosts := "localhost"
	if host != "" && host != "localhost" {
		hosts += "," + host
	}
	ips := []net.IP{net.ParseIP("127.0.0.1")}

	caCert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              "localhost",
		IPAddresses:       ips,
		Name:              "Kafka CA",
		SubjectCommonName: "Kafka CA",
		IsCA:              true,
	})
	if err != nil {
		return nil, fmt.Errorf("generate CA certificate: %w", err)
	}

	serverCert, err := tlscert.SelfSignedFromRequestE(tlscert.Request{
		Host:              hosts,
		IPAddresses:       ips,
		Name:              "Kafka Server",
		SubjectCommonName: "localhost",
		Parent:            caCert,
	})
	if err != nil {
		return nil, fmt.Errorf("generate server certificate: %w", err)
	}

	keystoreBytes, err := pkcs12.Modern.Encode(
		serverCert.Key,
		serverCert.Cert,
		[]*x509.Certificate{caCert.Cert},
		keystorePassword,
	)
	if err != nil {
		return nil, fmt.Errorf("encode PKCS12 keystore: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert.Cert)

	return &tlsCerts{
		KeystoreBytes: keystoreBytes,
		TLSConfig: &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		},
	}, nil
}
