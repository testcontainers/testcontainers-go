package cockroachdb

import (
	"context"
	"net"
	"net/url"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultSQLPort   = "26257/tcp"
	defaultAdminPort = "8080/tcp"

	defaultImage     = "cockroachdb/cockroach:latest-v23.1"
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

// MustConnectionString panics if the address cannot be determined.
func (c *CockroachDBContainer) MustConnectionString(ctx context.Context) string {
	addr, err := c.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}
	return addr
}

// ConnectionString returns the dial address to open a new connection to CockroachDB.
func (c *CockroachDBContainer) ConnectionString(ctx context.Context) (string, error) {
	port, err := c.MappedPort(ctx, defaultSQLPort)
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return connString(c.opts, host, port), nil
}

// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	o := defaultOptions()
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: defaultImage,
			ExposedPorts: []string{
				defaultSQLPort,
				defaultAdminPort,
			},
			WaitingFor: wait.ForAll(
				wait.ForHTTP("/health").WithPort(defaultAdminPort),
				wait.ForSQL(nat.Port(defaultSQLPort), "pgx/v5", func(host string, port nat.Port) string {
					return connString(o, host, port)
				}),
			),
		},
		Started: true,
	}

	// apply options
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&o)
		}
		opt.Customize(&req)
	}

	addCmd(&req, o)
	addEnvs(&req, o)

	// start
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}
	return &CockroachDBContainer{Container: container, opts: o}, nil
}

func addCmd(req *testcontainers.GenericContainerRequest, opts options) {
	req.Cmd = []string{
		"start-single-node",
		"--store=type=mem,size=" + opts.StoreSize,
	}

	if opts.Password == "" {
		req.Cmd = append(req.Cmd, "--insecure")
	} else {
		// password authentication
		req.Cmd = append(req.Cmd, "--accept-sql-without-tls")
	}
}

func addEnvs(req *testcontainers.GenericContainerRequest, opts options) {
	if req.Env == nil {
		req.Env = make(map[string]string)
	}

	req.Env["COCKROACH_DATABASE"] = opts.Database
	req.Env["COCKROACH_USER"] = opts.User
	req.Env["COCKROACH_PASSWORD"] = opts.Password
}

func connString(opts options, host string, port nat.Port) string {
	params := url.Values{
		"sslmode": []string{"disable"},
	}

	user := url.User(opts.User)
	if opts.Password != "" {
		user = url.UserPassword(opts.User, opts.Password)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port.Port()),
		Path:     opts.Database,
		RawQuery: params.Encode(),
	}

	return u.String()
}
