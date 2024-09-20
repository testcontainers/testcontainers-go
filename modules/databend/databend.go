package databend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	databendUser        = "databend"
	defaultUser         = "databend"
	defaultPassword     = "databend"
	defaultDatabaseName = "default"
)

// Container represents the Databend container type used in the module
type Container struct {
	*testcontainers.DockerContainer
	username string
	password string
	database string
}

var _ testcontainers.RequestCustomizer = (*DatabendOption)(nil)

// DatabendOption is an option for the Databend container.
type DatabendOption func(*Container)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o DatabendOption) Customize(*testcontainers.Request) error {
	// NOOP to satisfy interface.
	return nil
}

// Run creates an instance of the Databend container type
func Run(ctx context.Context, img string, opts ...testcontainers.RequestCustomizer) (*Container, error) {
	req := testcontainers.Request{
		Image:        img,
		ExposedPorts: []string{"8000/tcp"},
		Env: map[string]string{
			"QUERY_DEFAULT_USER":     defaultUser,
			"QUERY_DEFAULT_PASSWORD": defaultPassword,
		},
		WaitingFor: wait.ForListeningPort("8000/tcp"),
		Started:    true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	username := req.Env["QUERY_DEFAULT_USER"]
	password := req.Env["QUERY_DEFAULT_PASSWORD"]
	if password == "" && username == "" {
		return nil, errors.New("empty password and user")
	}

	container, err := testcontainers.Run(ctx, req)
	var c *Container
	if container != nil {
		c = &Container{
			DockerContainer: container,
			password:        password,
			username:        username,
			database:        defaultDatabaseName,
		}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

// MustConnectionString panics if the address cannot be determined.
func (c *Container) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8000/tcp")
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = "?" + strings.Join(args, "&")
	}
	if c.database == "" {
		return "", errors.New("database name is empty")
	}

	// databend://databend:databend@localhost:8000/default?sslmode=disable
	connectionString := fmt.Sprintf("databend://%s:%s@%s:%s/%s%s", c.username, c.password, host, containerPort.Port(), c.database, extraArgs)
	return connectionString, nil
}

// WithUsername sets the username for the Databend container.
// WithUsername is [Run] option that configures the default query user by setting
// the `QUERY_DEFAULT_USER` container environment variable.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
		req.Env["QUERY_DEFAULT_USER"] = username
		return nil
	}
}

// WithPassword sets the password for the Databend container.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
		req.Env["QUERY_DEFAULT_PASSWORD"] = password
		return nil
	}
}
