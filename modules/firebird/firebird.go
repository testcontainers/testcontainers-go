// Package firebird provides a Testcontainers module for running Firebird database containers.
package firebird

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultDatabase = "test.fdb"
	defaultUser     = "test"
	defaultPassword = "test"

	// dbPath is the directory inside the container where the jacobalberty/firebird
	// image stores database files. FIREBIRD_DATABASE is relative to this directory.
	dbPath = "/firebird/data"
)

// Container represents the Firebird container type used in the module.
type Container struct {
	testcontainers.Container
	database string
	user     string
	password string
}

// Run creates an instance of the Firebird container type.
// The default image is ghcr.io/jacobalberty/firebird:v3.0.
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

	// Retrieve credentials from the container environment so that option
	// overrides (WithDatabase, WithUsername, WithPassword) are reflected.
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect firebird: %w", err)
	}

	var foundDB, foundUser, foundPass bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "FIREBIRD_DATABASE="); ok {
			c.database, foundDB = v, true
		} else if v, ok := strings.CutPrefix(env, "FIREBIRD_USER="); ok {
			c.user, foundUser = v, true
		} else if v, ok := strings.CutPrefix(env, "FIREBIRD_PASSWORD="); ok {
			c.password, foundPass = v, true
		}

		if foundDB && foundUser && foundPass {
			break
		}
	}

	return c, nil
}

// ConnectionString returns the connection string for the Firebird container.
// It uses the firebird:// scheme and the full server-side path to the database
// file (/firebird/data/<name>), which is where the jacobalberty/firebird image
// creates the database. IPv6 hosts are handled correctly via net.JoinHostPort.
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	port, err := c.MappedPort(ctx, "3050/tcp")
	if err != nil {
		return "", err
	}

	// The jacobalberty/firebird image stores databases under /firebird/data/.
	// Using the full absolute path prevents "database not found" errors that
	// occur when clients receive only a bare filename.
	return fmt.Sprintf("firebird://%s:%s@%s%s/%s?charset=UTF8",
		c.user, c.password,
		net.JoinHostPort(host, port.Port()),
		dbPath,
		c.database), nil
}

// WithDatabase sets the FIREBIRD_DATABASE environment variable to configure the
// name of the database file to create or connect to.
func WithDatabase(name string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["FIREBIRD_DATABASE"] = name
		return nil
	}
}

// WithUsername sets the FIREBIRD_USER environment variable to configure the
// database user.
func WithUsername(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["FIREBIRD_USER"] = user
		return nil
	}
}

// WithPassword sets the FIREBIRD_PASSWORD environment variable to configure the
// password for the database user.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["FIREBIRD_PASSWORD"] = password
		return nil
	}
}

// WithSYSDBAPassword sets the ISC_PASSWORD environment variable to configure the
// SYSDBA master password.
func WithSYSDBAPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ISC_PASSWORD"] = password
		return nil
	}
}
