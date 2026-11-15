package examples_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"
	"strings"
	"testing"

	"github.com/mdelapenya/tlscert"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestRunContainer_connectUsingAmqp(t *testing.T) {
	ctx := context.Background()

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12.11-management-alpine")
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	require.NoError(t, err)

	amqpConnection, err := amqp.Dial(amqpURL)
	require.NoError(t, err)

	err = amqpConnection.Close()
	require.NoError(t, err)
}

func TestRunContainer_connectUsingAmqps(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()

	caCert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "ca",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		ParentDir: tmpDir,
	})
	require.NotNilf(t, caCert, "failed to generate CA certificate")

	cert := tlscert.SelfSignedFromRequest(tlscert.Request{
		Name:      "client",
		Host:      "localhost,127.0.0.1",
		IsCA:      true,
		Parent:    caCert,
		ParentDir: tmpDir,
	})
	require.NotNilf(t, cert, "failed to generate certificate")

	sslSettings := rabbitmq.SSLSettings{
		CACertFile:        caCert.CertPath,
		CertFile:          cert.CertPath,
		KeyFile:           cert.KeyPath,
		VerificationMode:  rabbitmq.SSLVerificationModePeer,
		FailIfNoCert:      false,
		VerificationDepth: 1,
	}

	rabbitmqContainer, err := rabbitmq.Run(ctx, "rabbitmq:3.12.11-management-alpine", rabbitmq.WithSSL(sslSettings))
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	amqpsURL, err := rabbitmqContainer.AmqpsURL(ctx)
	require.NoError(t, err)

	require.Truef(t, strings.HasPrefix(amqpsURL, "amqps"), "AMQPS Url should begin with `amqps`")

	certs := x509.NewCertPool()

	pemData, err := os.ReadFile(sslSettings.CACertFile)
	require.NoError(t, err)
	certs.AppendCertsFromPEM(pemData)

	amqpsConnection, err := amqp.DialTLS(amqpsURL, &tls.Config{InsecureSkipVerify: false, RootCAs: certs})
	require.NoError(t, err)

	require.Falsef(t, amqpsConnection.IsClosed(), "AMQPS Connection unexpectdely closed")
	err = amqpsConnection.Close()
	require.NoError(t, err)
}

func TestRunContainer_withCustomCredentials(t *testing.T) {
	ctx := context.Background()

	customUsername := "admin"
	customPassword := "s3cr3t"

	rabbitmqContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.12.11-management-alpine",
		rabbitmq.WithAdminUsername(customUsername),
		rabbitmq.WithAdminPassword(customPassword),
	)
	testcontainers.CleanupContainer(t, rabbitmqContainer)
	require.NoError(t, err)

	// Verify the container reports the custom credentials
	require.Equal(t, customUsername, rabbitmqContainer.AdminUsername)
	require.Equal(t, customPassword, rabbitmqContainer.AdminPassword)

	// Get the AMQP URL - this will include the custom credentials
	amqpURL, err := rabbitmqContainer.AmqpURL(ctx)
	require.NoError(t, err)
	require.Contains(t, amqpURL, customUsername)
	require.Contains(t, amqpURL, customPassword)

	// Try to connect using the URL with custom credentials
	amqpConnection, err := amqp.Dial(amqpURL)
	require.NoError(t, err)
	require.NoError(t, amqpConnection.Close())
}
