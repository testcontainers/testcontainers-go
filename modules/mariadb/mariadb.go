package mariadb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	rootUser            = "root"
	defaultUser         = "test"
	defaultPassword     = "test"
	defaultDatabaseName = "test"
)

// MariaDBContainer represents the MariaDB container type used in the module
type MariaDBContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MariaDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error) {
	return Run(ctx, "mariadb:11.0.3", opts...)
}

// Run creates an instance of the MariaDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MariaDBContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("3306/tcp", "33060/tcp"),
		testcontainers.WithEnv(map[string]string{
			"MARIADB_USER":     defaultUser,
			"MARIADB_PASSWORD": defaultPassword,
			"MARIADB_DATABASE": defaultDatabaseName,
		}),
		testcontainers.WithWaitStrategy(wait.ForLog("port: 3306  mariadb.org binary distribution")),
	}

	opts = append(opts, WithDefaultCredentials())

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("mariadb option: %w", err)
			}
		}
	}

	// Apply MySQL environment variables after user customization
	// In future releases of MariaDB, they could remove the MYSQL_* environment variables
	// at all. Then we can remove this customization.
	moduleOpts = append(moduleOpts, withMySQLEnvVars())

	username, ok := defaultOptions.env["MARIADB_USER"]
	if !ok {
		username = rootUser
	}
	password := defaultOptions.env["MARIADB_PASSWORD"]

	if len(password) == 0 && password == "" && !strings.EqualFold(rootUser, username) {
		return nil, errors.New("empty password can be used only with the root user")
	}

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MariaDBContainer
	if ctr != nil {
		c = &MariaDBContainer{
			Container: ctr,
			username:  username,
			password:  password,
			database:  defaultOptions.env["MARIADB_DATABASE"],
		}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// MustConnectionString panics if the address cannot be determined.
func (c *MariaDBContainer) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

func (c *MariaDBContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = strings.Join(args, "&")
	}
	if extraArgs != "" {
		extraArgs = "?" + extraArgs
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", c.username, c.password, host, containerPort.Port(), c.database, extraArgs)
	return connectionString, nil
}
