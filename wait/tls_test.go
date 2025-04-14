package wait_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	serverName         = "127.0.0.1"
	caFilename         = "/tmp/ca.pem"
	clientCertFilename = "/tmp/cert.crt"
	clientKeyFilename  = "/tmp/cert.key"
)

var (
	//go:embed testdata/cert.crt
	certBytes []byte

	//go:embed testdata/cert.key
	keyBytes []byte
)

// testForTLSCert creates a new CertStrategy for testing.
func testForTLSCert() *wait.TLSStrategy {
	return wait.ForTLSCert(clientCertFilename, clientKeyFilename).
		WithRootCAs(caFilename).
		WithServerName(serverName).
		WithStartupTimeout(time.Millisecond * 50).
		WithPollInterval(time.Millisecond)
}

func TestForCert(t *testing.T) {
	errNotFound := errdefs.NotFound(errors.New("file not found"))
	ctx := context.Background()

	t.Run("ca-not-found", func(t *testing.T) {
		target := newRunningTarget()
		target.EXPECT().CopyFileFromContainer(anyContext, caFilename).Return(nil, errNotFound)
		err := testForTLSCert().WaitUntilReady(ctx, target)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})

	t.Run("cert-not-found", func(t *testing.T) {
		target := newRunningTarget()
		caFile := io.NopCloser(bytes.NewBuffer(caBytes))
		target.EXPECT().CopyFileFromContainer(anyContext, caFilename).Return(caFile, nil)
		target.EXPECT().CopyFileFromContainer(anyContext, clientCertFilename).Return(nil, errNotFound)
		err := testForTLSCert().WaitUntilReady(ctx, target)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})

	t.Run("key-not-found", func(t *testing.T) {
		target := newRunningTarget()
		caFile := io.NopCloser(bytes.NewBuffer(caBytes))
		certFile := io.NopCloser(bytes.NewBuffer(certBytes))
		target.EXPECT().CopyFileFromContainer(anyContext, caFilename).Return(caFile, nil)
		target.EXPECT().CopyFileFromContainer(anyContext, clientCertFilename).Return(certFile, nil)
		target.EXPECT().CopyFileFromContainer(anyContext, clientKeyFilename).Return(nil, errNotFound)
		err := testForTLSCert().WaitUntilReady(ctx, target)
		require.EqualError(t, err, context.DeadlineExceeded.Error())
	})

	t.Run("valid", func(t *testing.T) {
		target := newRunningTarget()
		caFile := io.NopCloser(bytes.NewBuffer(caBytes))
		certFile := io.NopCloser(bytes.NewBuffer(certBytes))
		keyFile := io.NopCloser(bytes.NewBuffer(keyBytes))
		target.EXPECT().CopyFileFromContainer(anyContext, caFilename).Return(caFile, nil)
		target.EXPECT().CopyFileFromContainer(anyContext, clientCertFilename).Return(certFile, nil)
		target.EXPECT().CopyFileFromContainer(anyContext, clientKeyFilename).Return(keyFile, nil)

		certStrategy := testForTLSCert()
		err := certStrategy.WaitUntilReady(ctx, target)
		require.NoError(t, err)

		pool := x509.NewCertPool()
		require.True(t, pool.AppendCertsFromPEM(caBytes))
		cert, err := tls.X509KeyPair(certBytes, keyBytes)
		require.NoError(t, err)
		got := certStrategy.TLSConfig()
		require.Equal(t, serverName, got.ServerName)
		require.Equal(t, []tls.Certificate{cert}, got.Certificates)
		require.True(t, pool.Equal(got.RootCAs))
	})
}

func ExampleForTLSCert() {
	ctx := context.Background()

	// waitForTLSCert {
	// The file names passed to ForTLSCert are the paths where the files will
	// be copied to in the container as detailed by the Dockerfile.
	forCert := wait.ForTLSCert("/app/tls.pem", "/app/tls-key.pem").
		WithServerName("testcontainer.go.test")
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context: "testdata/http",
		},
		WaitingFor: forCert,
	}
	// }

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	state, err := c.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	// waitTLSConfig {
	config := forCert.TLSConfig()
	// }
	fmt.Println(config.ServerName)
	fmt.Println(len(config.Certificates))

	// Output:
	// true
	// testcontainer.go.test
	// 1
}
