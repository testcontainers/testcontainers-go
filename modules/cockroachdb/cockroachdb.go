package cockroachdb

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ErrTLSNotEnabled is returned when trying to get a TLS config from a container that does not have TLS enabled.
var ErrTLSNotEnabled = errors.New("tls not enabled")

const (
	defaultSQLPort   = "26257/tcp"
	defaultAdminPort = "8080/tcp"

	defaultUser      = "root"
	defaultPassword  = ""
	defaultDatabase  = "defaultdb"
	defaultStoreSize = "100%"

	// initDBPath is the path where the init scripts are placed in the container.
	initDBPath = "/docker-entrypoint-initdb.d"

	// cockroachDir is the path where the CockroachDB files are placed in the container.
	cockroachDir = "/cockroach"

	// clusterDefaultsContainerFile is the path to the default cluster settings script in the container.
	clusterDefaultsContainerFile = initDBPath + "/__cluster_defaults.sql"

	// memStorageFlag is the flag to use in the start command to use an in-memory store.
	memStorageFlag = "--store=type=mem,size="

	// insecureFlag is the flag to use in the start command to disable TLS.
	insecureFlag = "--insecure"

	// env vars.
	envUser     = "COCKROACH_USER"
	envPassword = "COCKROACH_PASSWORD"
	envDatabase = "COCKROACH_DATABASE"

	// cert files.
	certsDir   = cockroachDir + "/certs"
	fileCACert = certsDir + "/ca.crt"
)

//go:embed data/cluster_defaults.sql
var clusterDefaults []byte

// defaultsReader is a reader for the default settings scripts
// so that they can be identified and removed from the request.
type defaultsReader struct {
	*bytes.Reader
}

// newDefaultsReader creates a new reader for the default cluster settings script.
func newDefaultsReader(data []byte) *defaultsReader {
	return &defaultsReader{Reader: bytes.NewReader(data)}
}

// CockroachDBContainer represents the CockroachDB container type used in the module
type CockroachDBContainer struct {
	testcontainers.Container
	options
}

// options represents the options for the CockroachDBContainer type.
type options struct {
	database    string
	user        string
	password    string
	tlsStrategy *wait.TLSStrategy
}

// MustConnectionString returns a connection string to open a new connection to CockroachDB
// as described by [CockroachDBContainer.ConnectionString].
// It panics if an error occurs.
func (c *CockroachDBContainer) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionInfo contains a portable PostgreSQL URL and its in-memory TLS settings.
// URL alone does not include the container's CA or client certificate.
type ConnectionInfo struct {
	URL       string
	TLSConfig *tls.Config
}

// ConnectionString returns a PostgreSQL URL, not a registered pgx stdlib handle.
// For secure connections, use ConnectionInfo to also obtain the CA and client
// certificate, or use the pgx adapter in the examples child module.
func (c *CockroachDBContainer) ConnectionString(ctx context.Context) (string, error) {
	info, err := c.ConnectionInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("connection info: %w", err)
	}
	return info.URL, nil
}

// ConnectionInfo returns the connection URL and a clone of the TLS config.
// TLSConfig is nil in insecure mode. Callers must apply it to their client;
// the URL cannot encode the in-memory CA and client certificate.
func (c *CockroachDBContainer) ConnectionInfo(ctx context.Context) (ConnectionInfo, error) {
	port, err := c.MappedPort(ctx, defaultSQLPort)
	if err != nil {
		return ConnectionInfo{}, fmt.Errorf("mapped port: %w", err)
	}
	host, err := c.Host(ctx)
	if err != nil {
		return ConnectionInfo{}, fmt.Errorf("host: %w", err)
	}
	var cfg *tls.Config
	if original := c.tlsStrategy.TLSConfig(); original != nil {
		cfg = original.Clone()
		if cfg.RootCAs != nil {
			cfg.RootCAs = cfg.RootCAs.Clone()
		}
	}
	return ConnectionInfo{URL: c.connectionURL(host, port.Port(), cfg != nil).String(), TLSConfig: cfg}, nil
}

// TLSConfig returns config necessary to connect to CockroachDB over TLS.
// Returns [ErrTLSNotEnabled] if TLS is not enabled.
//
// Deprecated: use [CockroachDBContainer.ConnectionInfo] instead.
func (c *CockroachDBContainer) TLSConfig() (*tls.Config, error) {
	if cfg := c.tlsStrategy.TLSConfig(); cfg != nil {
		return cfg, nil
	}

	return nil, ErrTLSNotEnabled
}

// Deprecated: use Run instead.
// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	return Run(ctx, "cockroachdb/cockroach:latest-v23.1", opts...)
}

// Run start an instance of the CockroachDB container type using the given image and options.
//
// By default, the container will be configured with:
//   - Cluster: Single node
//   - Storage: 100% in-memory
//   - User: root
//   - Password: ""
//   - Database: defaultdb
//   - Exposed ports: 26257/tcp (SQL), 8080/tcp (Admin UI)
//   - Init Scripts: `data/cluster_defaults.sql`
//
// This supports CockroachDB images v22.2.0 and later, earlier versions will only work with
// customised options, such as disabling TLS and removing the wait for `init_success` using
// a [testcontainers.ContainerCustomizer].
//
// The init script `data/cluster_defaults.sql` configures the settings recommended
// by Cockroach Labs for [local testing clusters] unless data exists in the
// `/cockroach/cockroach-data` directory within the container. Use [WithNoClusterDefaults]
// to disable this behaviour and provide your own settings using [WithInitScripts].
//
// For more information see starting a [local cluster in docker].
//
// [local cluster in docker]: https://www.cockroachlabs.com/docs/stable/start-a-local-cluster-in-docker-linux
// [local testing clusters]: https://www.cockroachlabs.com/docs/stable/local-testing
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	ctr := &CockroachDBContainer{
		options: options{
			database: defaultDatabase,
			user:     defaultUser,
			password: defaultPassword,
		},
	}

	// The URL is filled after options are applied; ForExec retains this slice.
	sqlCommand := []string{"cockroach", "sql", "--url", "", "--execute", "SELECT 1"}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 5+len(opts)+1)
	moduleOpts = append(moduleOpts,
		testcontainers.WithCmd(
			"start-single-node",
			memStorageFlag+defaultStoreSize,
		),
		testcontainers.WithExposedPorts(defaultSQLPort, defaultAdminPort),
		testcontainers.WithEnv(map[string]string{
			"COCKROACH_DATABASE": defaultDatabase,
			"COCKROACH_USER":     defaultUser,
			"COCKROACH_PASSWORD": defaultPassword,
		}),
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            newDefaultsReader(clusterDefaults),
			ContainerFilePath: clusterDefaultsContainerFile,
			FileMode:          0o644,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForFile(cockroachDir+"/init_success"),
			wait.ForHTTP("/health").WithPort(defaultAdminPort),
			wait.ForTLSCert(
				certsDir+"/client."+defaultUser+".crt",
				certsDir+"/client."+defaultUser+".key",
			).WithRootCAs(fileCACert).WithServerName("127.0.0.1"),
			wait.ForExec(sqlCommand),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	// configure the wait strategy after all the options have been applied
	// It extracts the TLS strategy from the wait strategy and sets it on the container.
	moduleOpts = append(moduleOpts, ctr.configure(sqlCommand))

	var err error
	ctr.Container, err = testcontainers.Run(ctx, img, moduleOpts...)
	if err != nil {
		return ctr, fmt.Errorf("run cockroachdb: %w", err)
	}

	return ctr, nil
}

// connectionURL builds a URL without depending on a SQL driver.
func (c *CockroachDBContainer) connectionURL(host string, port string, secure bool) *url.URL {
	user := url.User(c.user)
	if c.password != "" {
		user = url.UserPassword(c.user, c.password)
	}
	sslMode := "disable"
	if secure {
		sslMode = "verify-full"
	}
	return &url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + c.database,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}
}
