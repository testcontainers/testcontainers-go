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

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

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

// ConnectionString returns a connection string to open a new connection to CockroachDB.
// The returned string is suitable for use by [sql.Open] but is not compatible with
// [pgx.ParseConfig], so if you want to call [pgx.ConnectConfig] use the
// [CockroachDBContainer.ConnectionConfig] method instead.
func (c *CockroachDBContainer) ConnectionString(ctx context.Context) (string, error) {
	cfg, err := c.ConnectionConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("connection config: %w", err)
	}

	return stdlib.RegisterConnConfig(cfg), nil
}

// ConnectionConfig returns a [pgx.ConnConfig] for the CockroachDB container.
// This can be passed to [pgx.ConnectConfig] to open a new connection.
func (c *CockroachDBContainer) ConnectionConfig(ctx context.Context) (*pgx.ConnConfig, error) {
	port, err := c.MappedPort(ctx, defaultSQLPort)
	if err != nil {
		return nil, fmt.Errorf("mapped port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("host: %w", err)
	}

	return c.connConfig(host, port)
}

// TLSConfig returns config necessary to connect to CockroachDB over TLS.
// Returns [ErrTLSNotEnabled] if TLS is not enabled.
//
// Deprecated: use [CockroachDBContainer.ConnectionString] or
// [CockroachDBContainer.ConnectionConfig] instead.
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
	c := &CockroachDBContainer{
		options: defaultOptions(),
	}

	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithCmd("start-single-node", memStorageFlag+defaultStoreSize),
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
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForFile(cockroachDir+"/init_success"),
			wait.ForHTTP("/health").WithPort(defaultAdminPort),
			wait.ForTLSCert(
				certsDir+"/client."+defaultUser+".crt",
				certsDir+"/client."+defaultUser+".key",
			).WithRootCAs(fileCACert).WithServerName("127.0.0.1"),
			wait.ForSQL(defaultSQLPort, "pgx/v5", func(host string, port nat.Port) string {
				connStr, err := c.connString(host, port)
				if err != nil {
					panic(err)
				}
				return connStr
			}),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&c.options); err != nil {
				return nil, fmt.Errorf("cockroachdb option: %w", err)
			}
		}
	}

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(c.env))
	moduleOpts = append(moduleOpts, configure(&c.options))

	// pass it last to make sure all options have been set.
	moduleOpts = append(moduleOpts, validatePassword())

	var err error
	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	c.Container = ctr

	return c, nil
}

// connString returns a connection string for the given host, port and options.
func (c *CockroachDBContainer) connString(host string, port nat.Port) (string, error) {
	cfg, err := c.connConfig(host, port)
	if err != nil {
		return "", fmt.Errorf("connection config: %w", err)
	}

	return stdlib.RegisterConnConfig(cfg), nil
}

// connConfig returns a [pgx.ConnConfig] for the given host, port and options.
func (c *CockroachDBContainer) connConfig(host string, port nat.Port) (*pgx.ConnConfig, error) {
	var user *url.Userinfo
	if c.env[envPassword] != "" {
		user = url.UserPassword(c.env[envUser], c.env[envPassword])
	} else {
		user = url.User(c.env[envUser])
	}

	sslMode := "disable"
	tlsConfig := c.tlsStrategy.TLSConfig()
	if tlsConfig != nil {
		sslMode = "verify-full"
	}
	params := url.Values{
		"sslmode": []string{sslMode},
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port.Port()),
		Path:     c.env[envDatabase],
		RawQuery: params.Encode(),
	}

	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.TLSConfig = tlsConfig

	return cfg, nil
}
