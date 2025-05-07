package cassandra

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

// TLSConfig represents the TLS configuration for Cassandra
type TLSConfig struct {
	KeystorePath    string
	CertificatePath string
	Config          *tls.Config
}

// Options represents the configuration options for the Cassandra container
type Options struct {
	tlsConfig *TLSConfig
}

// Option is an option for the Cassandra container.
type Option func(*testcontainers.GenericContainerRequest, *Options) error

// Customize implements the testcontainers.ContainerCustomizer interface
func (o Option) Customize(req *testcontainers.GenericContainerRequest) error {
	return o(req, &Options{})
}

// WithSSL enables SSL/TLS support on the Cassandra container
func WithSSL() Option {
	return func(req *testcontainers.GenericContainerRequest, settings *Options) error {
		req.ExposedPorts = append(req.ExposedPorts, string(securePort))

		keystorePath, certPath, err := GenerateJKSKeystore()
		if err != nil {
			return fmt.Errorf("create SSL certs: %w", err)
		}

		req.Files = append(req.Files,
			testcontainers.ContainerFile{
				HostFilePath:      keystorePath,
				ContainerFilePath: "/etc/cassandra/conf/keystore.jks",
				FileMode:          0o644,
			},
			testcontainers.ContainerFile{
				HostFilePath:      certPath,
				ContainerFilePath: "/etc/cassandra/conf/cassandra.crt",
				FileMode:          0o644,
			})

		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("error while read certificate: %w", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(certPEM) {
			return fmt.Errorf("failed to append certificate to pool")
		}

		settings.tlsConfig = &TLSConfig{
			KeystorePath:    keystorePath,
			CertificatePath: certPath,
			Config: &tls.Config{
				RootCAs:    certPool,
				ServerName: "localhost",
				MinVersion: tls.VersionTLS12,
			},
		}

		return nil
	}
}

// GenerateJKSKeystore generates a JKS keystore with a self-signed cert using keytool, and extracts the public cert for Go client trust.
func GenerateJKSKeystore() (keystorePath, certPath string, err error) {
	tmpDir := os.TempDir()
	keystorePath = filepath.Join(tmpDir, "keystore.jks")
	keystorePassword := "changeit"
	certPath = filepath.Join(tmpDir, "cert.pem")

	// Remove existing keystore if it exists
	os.Remove(keystorePath)

	// Generate keystore with self-signed certificate
	cmd := exec.Command(
		"keytool", "-genkeypair",
		"-alias", "cassandra",
		"-keyalg", "RSA",
		"-keysize", "2048",
		"-storetype", "JKS",
		"-keystore", keystorePath,
		"-storepass", keystorePassword,
		"-keypass", keystorePassword,
		"-dname", "CN=localhost, OU=Test, O=Test, C=US",
		"-validity", "365",
	)
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to generate keystore: %w", err)
	}

	// Export the public certificate for Go client trust
	cmd = exec.Command(
		"keytool", "-exportcert",
		"-alias", "cassandra",
		"-keystore", keystorePath,
		"-storepass", keystorePassword,
		"-rfc",
		"-file", certPath,
	)
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("failed to export certificate: %w", err)
	}

	return keystorePath, certPath, nil
}

// TLSConfig returns the TLS configuration
func (o *Options) TLSConfig() *TLSConfig {
	return o.tlsConfig
}
