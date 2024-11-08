package cockroachdb

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"errors"
	"fmt"
	"io"
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

// options represents the options for the CockroachDBContainer type.
type options struct {
	// Settings.
	database string
	user     string
	password string
	insecure bool

	// Client certificate.
	clientCert []byte
	clientKey  []byte
	certPool   *x509.CertPool
	tlsConfig  *tls.Config
}

// WaitUntilReady implements the [wait.Strategy] interface.
// If TLS is enabled, it waits for the CA, client cert and key for the configured user to be
// available in the container and uses them to setup the TLS config, otherwise it does nothing.
//
// This is defined on the options as it needs to know the customised values to operate correctly.
func (o *options) WaitUntilReady(ctx context.Context, target wait.StrategyTarget) error {
	if o.insecure {
		return nil
	}

	return wait.ForAll(
		wait.ForFile(fileCACert).WithMatcher(func(r io.Reader) error {
			buf, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("read CA cert: %w", err)
			}

			if !o.certPool.AppendCertsFromPEM(buf) {
				return errors.New("invalid CA cert")
			}

			return nil
		}),
		wait.ForFile(certsDir+"/client."+o.user+".crt").WithMatcher(func(r io.Reader) error {
			var err error
			if o.clientCert, err = io.ReadAll(r); err != nil {
				return fmt.Errorf("read client cert: %w", err)
			}

			return nil
		}),
		wait.ForFile(certsDir+"/client."+o.user+".key").WithMatcher(func(r io.Reader) error {
			var err error
			if o.clientKey, err = io.ReadAll(r); err != nil {
				return fmt.Errorf("read client key: %w", err)
			}

			cert, err := tls.X509KeyPair(o.clientCert, o.clientKey)
			if err != nil {
				return fmt.Errorf("x509 key pair: %w", err)
			}

			o.tlsConfig = &tls.Config{
				RootCAs:      o.certPool,
				Certificates: []tls.Certificate{cert},
				ServerName:   "127.0.0.1",
			}

			return nil
		}),
	).WaitUntilReady(ctx, target)
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
// The returned string is suitable for use by [sql.Open] but is not be compatible with
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
// Deprecated: use [CockroachDBContainer.ConnectionConfig] or
// [CockroachDBContainer.ConnectionConfig] instead.
func (c *CockroachDBContainer) TLSConfig() (*tls.Config, error) {
	if c.tlsConfig == nil {
		return nil, ErrTLSNotEnabled
	}

	return c.tlsConfig, nil
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
	if c.password != "" {
		user = url.UserPassword(c.user, c.password)
	} else {
		user = url.User(c.user)
	}

	sslMode := "disable"
	if c.tlsConfig != nil {
		sslMode = "verify-full"
	}
	params := url.Values{
		"sslmode": []string{sslMode},
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port.Port()),
		Path:     c.database,
		RawQuery: params.Encode(),
	}

	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.TLSConfig = c.tlsConfig

	return cfg, nil
}

// setOptions sets the CockroachDBContainer options from a request.
func (c *CockroachDBContainer) setOptions(req *testcontainers.GenericContainerRequest) {
	c.database = req.Env[envDatabase]
	c.user = req.Env[envUser]
	c.password = req.Env[envPassword]
	for _, arg := range req.Cmd {
		if arg == insecureFlag {
			c.insecure = true
			break
		}
	}
}

// Deprecated: use Run instead.
// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	return Run(ctx, "cockroachdb/cockroach:latest-v23.1", opts...)
}

// Run start an instance of the CockroachDB container type using the given image and options.
//
// By default, the container will:
//   - Cluster: Single node
//   - Storage: 100% in-memory
//   - User: root
//   - Password: ""
//   - Database: defaultdb
//   - Exposed ports: 26257/tcp (SQL), 8080/tcp (Admin UI)
//
// For more information see starting a [local cluster in docker].
//
// [local cluster in docker]: https://www.cockroachlabs.com/docs/stable/start-a-local-cluster-in-docker-linux
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	ctr := &CockroachDBContainer{
		options: options{
			database: defaultDatabase,
			user:     defaultUser,
			password: defaultPassword,
			certPool: x509.NewCertPool(),
		},
	}
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: img,
			ExposedPorts: []string{
				defaultSQLPort,
				defaultAdminPort,
			},
			Env: map[string]string{
				"COCKROACH_DATABASE": defaultDatabase,
				"COCKROACH_USER":     defaultUser,
				"COCKROACH_PASSWORD": defaultPassword,
			},
			Files: []testcontainers.ContainerFile{{
				Reader:            newDefaultsReader(clusterDefaults),
				ContainerFilePath: clusterDefaultsContainerFile,
				FileMode:          0o644,
			}},
			Cmd: []string{
				"start-single-node",
				memStorageFlag + defaultStoreSize,
			},
			WaitingFor: wait.ForAll(
				wait.ForFile(cockroachDir+"/init_success"),
				wait.ForHTTP("/health").WithPort(defaultAdminPort),
				ctr, // Wait for the TLS files to be available if needed.
				wait.ForSQL(defaultSQLPort, "pgx/v5", func(host string, port nat.Port) string {
					connStr, err := ctr.connString(host, port)
					if err != nil {
						panic(err)
					}
					return connStr
				}),
			),
		},
		Started: true,
	}

	for _, opt := range opts {
		if fn, ok := opt.(customizer); ok {
			if err := fn.customize(ctr); err != nil {
				return nil, fmt.Errorf("customize container: %w", err)
			}
		}
		if err := opt.Customize(&req); err != nil {
			return nil, fmt.Errorf("customize request: %w", err)
		}
	}

	// Extract the options from the request so they can used by wait strategies and connection methods.
	ctr.setOptions(&req)

	var err error
	ctr.Container, err = testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return ctr, fmt.Errorf("generic container: %w", err)
	}

	return ctr, nil
}
