// Package cubrid provides a Testcontainers module for CUBRID.
package cubrid

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	brokerPort = "33000/tcp"

	defaultDatabase = "testdb"
	defaultUser     = "testuser"
	defaultPassword = "testpass"
)

// Container represents the CUBRID container type used in the module.
type Container struct {
	testcontainers.Container
	database string
	user     string
	password string
}

// Run creates an instance of the CUBRID container type.
// It returns once the broker and the database server have started, and fails at once
// with the error lines from the container output when CUBRID cannot start.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 4+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(brokerPort),
		testcontainers.WithEnv(map[string]string{
			"CUBRID_DB":       defaultDatabase,
			"CUBRID_USER":     defaultUser,
			"CUBRID_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategyAndDeadline(startupTimeout,
			startupStrategy(),
			wait.ForListeningPort(brokerPort),
		),
	)
	moduleOpts = append(moduleOpts, opts...)
	moduleOpts = append(moduleOpts, testcontainers.CustomizeRequestOption(validateSettings))

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
		return c, fmt.Errorf("run cubrid: %w", err)
	}

	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect cubrid: %w", err)
	}

	var foundDB, foundUser, foundPass bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "CUBRID_DB="); ok {
			c.database, foundDB = v, true
		} else if v, ok := strings.CutPrefix(env, "CUBRID_USER="); ok {
			c.user, foundUser = v, true
		} else if v, ok := strings.CutPrefix(env, "CUBRID_PASSWORD="); ok {
			c.password, foundPass = v, true
		}

		if foundDB && foundUser && foundPass {
			break
		}
	}

	return c, nil
}

// ConnectionString returns the URL of the database, cubrid://user:password@host:port/database,
// which is the form used by pure Go CUBRID drivers.
// Each of args is appended as a query parameter, such as "autocommit=false".
func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, brokerPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	u := url.URL{
		Scheme:   "cubrid",
		User:     url.UserPassword(c.user, c.password),
		Host:     endpoint,
		Path:     "/" + c.database,
		RawQuery: strings.Join(args, "&"),
	}
	return u.String(), nil
}
