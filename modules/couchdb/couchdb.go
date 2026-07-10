package couchdb

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort     = "5984/tcp"
	defaultUser     = "admin"
	defaultPassword = "password"
)

// Container represents the CouchDB container type used in the module
type Container struct {
	testcontainers.Container
	user     string
	password string
}

// Run creates an instance of the CouchDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"COUCHDB_USER":     defaultUser,
			"COUCHDB_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/_up").
				WithPort(defaultPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				}),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, user: defaultUser, password: defaultPassword}
	}

	if err != nil {
		return c, fmt.Errorf("run couchdb: %w", err)
	}

	// Inspect the container to get the actual credentials from environment
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect couchdb: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "COUCHDB_USER="); ok {
			c.user = v
		}
		if v, ok := strings.CutPrefix(env, "COUCHDB_PASSWORD="); ok {
			c.password = v
		}
	}

	return c, nil
}

// WithAdminCredentials sets the admin username and password for the CouchDB container.
func WithAdminCredentials(user, password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["COUCHDB_USER"] = user
		req.Env["COUCHDB_PASSWORD"] = password
		return nil
	}
}

// ConnectionString returns the connection string for the CouchDB container,
// using the format: http://user:password@host:5984
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s@%s:%s", c.user, c.password, host, port.Port()), nil
}
