package firebird

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultDatabase = "test.fdb"
	defaultUser     = "test"
	defaultPassword = "test"
)

// Container represents the Firebird container type used in the module
type Container struct {
	testcontainers.Container
	database string
	user     string
	password string
}

// Run creates an instance of the Firebird container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts("3050/tcp"),
		testcontainers.WithEnv(map[string]string{
			"FIREBIRD_DATABASE": defaultDatabase,
			"FIREBIRD_USER":     defaultUser,
			"FIREBIRD_PASSWORD": defaultPassword,
			"ISC_PASSWORD":      "masterkey",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("3050/tcp").WithStartupTimeout(60*time.Second),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{
			Container: ctr,
			database:  defaultDatabase,
			user:      defaultUser,
			password:  defaultPassword,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run firebird: %w", err)
	}

	// Retrieve credentials from container environment
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect firebird: %w", err)
	}

	var foundDB, foundUser, foundPass bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "FIREBIRD_DATABASE="); ok {
			c.database, foundDB = v, true
		}
		if v, ok := strings.CutPrefix(env, "FIREBIRD_USER="); ok {
			c.user, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "FIREBIRD_PASSWORD="); ok {
			c.password, foundPass = v, true
		}

		if foundDB && foundUser && foundPass {
			break
		}
	}

	return c, nil
}

// ConnectionString returns the connection string for the Firebird container using
// the firebird:// scheme. Credentials are URL-encoded to handle special characters.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, "3050/tcp")
	if err != nil {
		return "", err
	}

	u := &url.URL{
		Scheme:   "firebird",
		User:     url.UserPassword(c.user, c.password),
		Host:     host + ":" + port.Port(),
		Path:     "/" + c.database,
		RawQuery: "charset=UTF8",
	}
	return u.String(), nil
}

// WithDatabase sets the FIREBIRD_DATABASE environment variable.
func WithDatabase(name string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["FIREBIRD_DATABASE"] = name
		return nil
	}
}

// WithUsername sets the FIREBIRD_USER environment variable.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["FIREBIRD_USER"] = user
		return nil
	}
}

// WithPassword sets the FIREBIRD_PASSWORD environment variable.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["FIREBIRD_PASSWORD"] = password
		return nil
	}
}

// WithSYSDBAPassword sets the ISC_PASSWORD environment variable (SYSDBA master password).
func WithSYSDBAPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["ISC_PASSWORD"] = password
		return nil
	}
}
