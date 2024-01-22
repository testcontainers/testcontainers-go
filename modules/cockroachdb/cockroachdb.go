package cockroachdb

import (
	"context"
	"net"
	"net/url"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultSQLPort   = "26257"
	defaultAdminPort = "8080"

	defaultImage     = "cockroachdb/cockroach:latest-v23.1"
	defaultDatabase  = "defaultdb"
	defaultStoreSize = "100%"
)

// CockroachDBContainer represents the CockroachDB container type used in the module
type CockroachDBContainer struct {
	testcontainers.Container
	opts Options
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
	mappedport, err := c.MappedPort(ctx, defaultSQLPort+"/tcp")
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"sslmode": []string{"disable"},
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.User("root"),
		Host:     net.JoinHostPort(hostIP, mappedport.Port()),
		Path:     c.opts.Database,
		RawQuery: params.Encode(),
	}

	return u.String(), nil
}

// RunContainer creates an instance of the CockroachDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*CockroachDBContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: defaultImage,
			ExposedPorts: []string{
				defaultSQLPort + "/tcp",
				defaultAdminPort + "/tcp",
			},
			WaitingFor: wait.ForHTTP("/health").WithPort(defaultAdminPort),
		},
		Started: true,
	}

	// apply options
	o := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&o)
		}
		opt.Customize(&req)
	}

	req.Cmd = cmd(o)

	// start
	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}
	return &CockroachDBContainer{Container: container, opts: o}, nil
}

func cmd(opts Options) []string {
	return []string{
		"start-single-node",
		"--insecure",
		"--store=type=mem,size=" + opts.StoreSize,
	}
}
