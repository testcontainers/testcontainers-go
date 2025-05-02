package cassandra_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/cassandra"
)

type Test struct {
	ID   uint64
	Name string
}

func TestCassandra(t *testing.T) {
	ctx := context.Background()

	ctr, err := cassandra.Run(ctx, "cassandra:4.1.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// connectionString {
	connectionHost, err := ctr.ConnectionHost(ctx)
	// }
	require.NoError(t, err)

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	defer session.Close()

	// perform assertions
	err = session.Query("CREATE KEYSPACE test_keyspace WITH REPLICATION = {'class' : 'SimpleStrategy', 'replication_factor' : 1}").Exec()
	require.NoError(t, err)
	err = session.Query("CREATE TABLE test_keyspace.test_table (id int PRIMARY KEY, name text)").Exec()
	require.NoError(t, err)

	err = session.Query("INSERT INTO test_keyspace.test_table (id, name) VALUES (1, 'NAME')").Exec()
	require.NoError(t, err)

	var test Test
	err = session.Query("SELECT id, name FROM test_keyspace.test_table WHERE id=1").Scan(&test.ID, &test.Name)
	require.NoError(t, err)
	require.Equal(t, Test{ID: 1, Name: "NAME"}, test)
}

func TestCassandraWithConfigFile(t *testing.T) {
	ctx := context.Background()

	ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithConfigFile(filepath.Join("testdata", "config.yaml")))
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	connectionHost, err := ctr.ConnectionHost(ctx)
	require.NoError(t, err)

	cluster := gocql.NewCluster(connectionHost)
	session, err := cluster.CreateSession()
	require.NoError(t, err)
	defer session.Close()

	var result string
	err = session.Query("SELECT cluster_name FROM system.local").Scan(&result)
	require.NoError(t, err)
	require.Equal(t, "My Cluster", result)
}

func TestCassandraWithInitScripts(t *testing.T) {
	t.Run("with init cql script", func(t *testing.T) {
		ctx := context.Background()

		// withInitScripts {
		ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithInitScripts(filepath.Join("testdata", "init.cql")))
		// }
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		// connectionHost {
		connectionHost, err := ctr.ConnectionHost(ctx)
		// }
		require.NoError(t, err)

		cluster := gocql.NewCluster(connectionHost)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var test Test
		err = session.Query("SELECT id, name FROM test_keyspace.test_table WHERE id=1").Scan(&test.ID, &test.Name)
		require.NoError(t, err)
		require.Equal(t, Test{ID: 1, Name: "NAME"}, test)
	})

	t.Run("with init bash script", func(t *testing.T) {
		ctx := context.Background()

		ctr, err := cassandra.Run(ctx, "cassandra:4.1.3", cassandra.WithInitScripts(filepath.Join("testdata", "init.sh")))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		connectionHost, err := ctr.ConnectionHost(ctx)
		require.NoError(t, err)

		cluster := gocql.NewCluster(connectionHost)
		session, err := cluster.CreateSession()
		require.NoError(t, err)
		defer session.Close()

		var test Test
		err = session.Query("SELECT id, name FROM init_sh_keyspace.test_table WHERE id=1").Scan(&test.ID, &test.Name)
		require.NoError(t, err)
		require.Equal(t, Test{ID: 1, Name: "NAME"}, test)
	})
}

func TestCassandraSSL(t *testing.T) {
	if _, err := exec.LookPath("keytool"); err != nil {
		t.Skip("keytool not found, skipping test")
	}

	ctx := context.Background()

	tmpDir := t.TempDir()
	keystorePath := filepath.Join(tmpDir, "keystore.jks")
	keystorePassword := "changeit"

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
	_, err := cmd.CombinedOutput()
	require.NoError(t, err)

	certPath := filepath.Join(tmpDir, "cassandra.crt")
	cmd = exec.Command(
		"keytool", "-exportcert",
		"-alias", "cassandra",
		"-keystore", keystorePath,
		"-storepass", keystorePassword,
		"-rfc",
		"-file", certPath,
	)
	_, err = cmd.CombinedOutput()
	require.NoError(t, err)

	sslOptions := cassandra.SSLOptions{
		KeystorePath:      keystorePath,
		CertPath:          certPath,
		RequireClientAuth: false,
	}

	container, err := cassandra.Run(ctx, "cassandra:4.1.3",
		// with ssl config file
		cassandra.WithConfigFile(filepath.Join("testdata", "cassandra-ssl.yaml")),
		cassandra.WithSSL(sslOptions))
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	sslPort, err := container.MappedPort(ctx, "9142/tcp")
	require.NoError(t, err)

	// Read the certificate for client validation
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatalf("Failed to read certificate: %v", err)
	}

	// Create a certificate pool and add the certificate
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)

	// Set up TLS configuration
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
	var session *gocql.Session
	session, err = cluster.CreateSession()
	require.NoError(t, err)
	defer session.Close()
	var version string
	err = session.Query("SELECT release_version FROM system.local").Scan(&version)
	require.NoError(t, err)
	require.NotEmpty(t, version)
}
