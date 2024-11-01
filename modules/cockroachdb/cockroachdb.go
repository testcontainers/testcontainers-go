package cockroachdb

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ErrTLSNotEnabled is returned when trying to get a TLS config from a container that does not have TLS enabled.
var ErrTLSNotEnabled = errors.New("tls not enabled")

const (
	certsDir = "/tmp"

	defaultSQLPort   = "26257/tcp"
	defaultAdminPort = "8080/tcp"

	defaultUser      = "root"
	defaultPassword  = ""
	defaultDatabase  = "defaultdb"
	defaultStoreSize = "100%"
)

// CockroachDBContainer represents the CockroachDB container type used in the module
type CockroachDBContainer struct {
	testcontainers.Container
	opts options
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
	return c.opts.containerConnString(ctx, c.Container)
}

// ConnectionConfig returns a [pgx.ConnConfig] for the CockroachDB container.
// This can be passed to [pgx.ConnectConfig] to open a new connection.
func (c *CockroachDBContainer) ConnectionConfig(ctx context.Context) (*pgx.ConnConfig, error) {
	return c.opts.containerConnConfig(ctx, c.Container)
}

// TLSConfig returns config necessary to connect to CockroachDB over TLS.
//
// Deprecated: use [CockroachDBContainer.ConnectionConfig] or
// [CockroachDBContainer.ConnectionConfig] instead.
func (c *CockroachDBContainer) TLSConfig() (*tls.Config, error) {
	if c.opts.TLS == nil {
		return nil, ErrTLSNotEnabled
	}

	return c.opts.TLS.tlsConfig()
}

// Deprecated: use Run instead.
// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	return Run(ctx, "cockroachdb/cockroach:latest-v23.1", opts...)
}

// Run creates an instance of the CockroachDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	o := defaultOptions()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: img,
			ExposedPorts: []string{
				defaultSQLPort,
				defaultAdminPort,
			},
			LifecycleHooks: []testcontainers.ContainerLifecycleHooks{
				{
					PreStarts: []testcontainers.ContainerHook{
						func(ctx context.Context, container testcontainers.Container) error {
							return addTLS(ctx, container, o)
						},
					},
					PostReadies: []testcontainers.ContainerHook{
						func(ctx context.Context, container testcontainers.Container) error {
							return runStatements(ctx, container, o)
						},
					},
				},
			},
		},
		Started: true,
	}

	// apply options
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&o)
		}
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	// modify request
	for _, fn := range []modiferFunc{
		addEnvs,
		addCmd,
		addWaitingFor,
	} {
		if err := fn(&req, o); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	var c *CockroachDBContainer
	if container != nil {
		c = &CockroachDBContainer{Container: container, opts: o}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

type modiferFunc func(*testcontainers.GenericContainerRequest, options) error

func addCmd(req *testcontainers.GenericContainerRequest, opts options) error {
	req.Cmd = []string{
		"start-single-node",
		"--store=type=mem,size=" + opts.StoreSize,
	}

	// authN
	if opts.TLS != nil {
		if opts.User != defaultUser {
			return fmt.Errorf("unsupported user %s with TLS, use %s", opts.User, defaultUser)
		}
		if opts.Password != "" {
			return errors.New("cannot use password authentication with TLS")
		}
	}

	switch {
	case opts.TLS != nil:
		req.Cmd = append(req.Cmd, "--certs-dir="+certsDir)
	case opts.Password != "":
		req.Cmd = append(req.Cmd, "--accept-sql-without-tls")
	default:
		req.Cmd = append(req.Cmd, "--insecure")
	}
	return nil
}

func addEnvs(req *testcontainers.GenericContainerRequest, opts options) error {
	if req.Env == nil {
		req.Env = make(map[string]string)
	}

	req.Env["COCKROACH_DATABASE"] = opts.Database
	req.Env["COCKROACH_USER"] = opts.User
	req.Env["COCKROACH_PASSWORD"] = opts.Password
	return nil
}

func addWaitingFor(req *testcontainers.GenericContainerRequest, opts options) error {
	sqlWait := wait.ForSQL(defaultSQLPort, "pgx/v5", func(host string, port nat.Port) string {
		connStr, err := opts.connString(host, port)
		if err != nil {
			panic(err)
		}
		return connStr
	})
	defaultStrategy := wait.ForAll(
		wait.ForHTTP("/health").WithPort(defaultAdminPort),
		sqlWait,
	)

	if req.WaitingFor == nil {
		req.WaitingFor = defaultStrategy
	} else {
		req.WaitingFor = wait.ForAll(req.WaitingFor, defaultStrategy)
	}

	return nil
}

func addTLS(ctx context.Context, container testcontainers.Container, opts options) error {
	if opts.TLS == nil {
		return nil
	}

	caBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: opts.TLS.CACert.Raw,
	})
	files := map[string][]byte{
		"ca.crt":          caBytes,
		"node.crt":        opts.TLS.NodeCert,
		"node.key":        opts.TLS.NodeKey,
		"client.root.crt": opts.TLS.ClientCert,
		"client.root.key": opts.TLS.ClientKey,
	}
	for filename, contents := range files {
		if err := container.CopyToContainer(ctx, contents, filepath.Join(certsDir, filename), 0o600); err != nil {
			return err
		}
	}
	return nil
}

// runStatements runs the configured statements against the CockroachDB container.
func runStatements(ctx context.Context, container testcontainers.Container, opts options) (err error) {
	if len(opts.Statements) == 0 {
		return nil
	}

	connStr, err := opts.containerConnString(ctx, container)
	if err != nil {
		return fmt.Errorf("connection string: %w", err)
	}

	db, err := sql.Open("pgx/v5", connStr)
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}
	defer func() {
		cerr := db.Close()
		if err == nil {
			err = cerr
		}
	}()

	for _, stmt := range opts.Statements {
		if _, err = db.Exec(stmt); err != nil {
			return fmt.Errorf("db.Exec: %w", err)
		}
	}

	return nil
}
