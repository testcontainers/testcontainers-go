package cassandra

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

// Options represents the configuration options for the Cassandra container
type Options struct {
	IsTLSEnabled bool
	TLSConfig    *tls.Config
}

// TLSEnabled returns whether TLS is enabled
func (o *Options) TLSEnabled() bool {
	return o.IsTLSEnabled
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Cassandra container.
type Option func(*Options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithSSL enables SSL/TLS support on the Cassandra container
func WithSSL() Option {
	return func(o *Options) error {
		o.IsTLSEnabled = true
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
