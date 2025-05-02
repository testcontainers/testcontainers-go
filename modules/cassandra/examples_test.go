package cassandra_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gocql/gocql"

	"crypto/tls"
	"crypto/x509"
	"os/exec"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

func ExampleRun() {
	// runCassandraContainer {
	ctx := context.Background()

	cassandraContainer, err := cassandra.Run(ctx,
		"cassandra:4.1.3",
		cassandra.WithInitScripts(filepath.Join("testdata", "init.cql")),
		cassandra.WithConfigFile(filepath.Join("testdata", "config.yaml")),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cassandraContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}
	// }

	state, err := cassandraContainer.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return
	}

	fmt.Println(state.Running)

	connectionHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Printf("failed to get connection host: %s", err)
		return
	}

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}
	defer session.Close()

	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	if err != nil {
		log.Printf("failed to query: %s", err)
		return
	}

	fmt.Println(version)

	// Output:
	// true
	// 4.1.3
}

func ExampleRunSSL() {
	ctx := context.Background()

	// Generate JKS keystore and export public cert for Go client
	tmpDir, err := os.MkdirTemp("", "cassandratestssl")
	if err != nil {
		log.Printf("failed to create temp dir: %s", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	keystorePath := filepath.Join(tmpDir, "keystore.jks")
	keystorePassword := "changeit"
	certPath := filepath.Join(tmpDir, "cert.pem")

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
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("failed to generate keystore: %s\n%s", err, string(out))
		return
	}

	cmd = exec.Command(
		"keytool", "-exportcert",
		"-alias", "cassandra",
		"-keystore", keystorePath,
		"-storepass", keystorePassword,
		"-rfc",
		"-file", certPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("failed to export cert: %s\n%s", err, string(out))
		return
	}

	cassandraContainer, err := cassandra.Run(ctx,
		"cassandra:4.1.3",
		cassandra.WithConfigFile(filepath.Join("testdata", "cassandra-ssl.yaml")),
		cassandra.WithSSL(cassandra.SSLOptions{
			KeystorePath:      keystorePath,
			KeystorePassword:  keystorePassword,
			CertPath:          certPath,
			RequireClientAuth: false,
		}),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(cassandraContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return
	}

	host, err := cassandraContainer.Host(ctx)
	if err != nil {
		log.Printf("failed to get host: %s", err)
		return
	}

	sslPort, err := cassandraContainer.MappedPort(ctx, "9142/tcp")
	if err != nil {
		log.Printf("failed to get SSL port: %s", err)
		return
	}

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		log.Printf("failed to read cert: %s", err)
		return
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)
	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: true, // For testing only
		ServerName:         "localhost",
		MinVersion:         tls.VersionTLS12,
	}

	cluster := gocql.NewCluster(fmt.Sprintf("%s:%s", host, sslPort.Port()))
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 30 * time.Second
	cluster.ConnectTimeout = 30 * time.Second
	cluster.DisableInitialHostLookup = true
	cluster.SslOpts = &gocql.SslOptions{
		Config:                 tlsConfig,
		EnableHostVerification: false,
	}
	session, err := cluster.CreateSession()
	if err != nil {
		log.Printf("failed to create session: %s", err)
		return
	}
	defer session.Close()

	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	if err != nil {
		log.Printf("failed to query: %s", err)
		return
	}

	fmt.Println(version)
	// Output:
	// 4.1.3
}
