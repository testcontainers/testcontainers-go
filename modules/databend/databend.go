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

// DatabendContainer represents the Databend container type used in the module
type DatabendContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

var _ testcontainers.ContainerCustomizer = (*DatabendOption)(nil)

// DatabendOption is an option for the Databend container.
type DatabendOption func(*DatabendContainer)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o DatabendOption) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// Run creates an instance of the Databend container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DatabendContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("8000/tcp")),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("databend option: %w", err)
			}
		}
	}

	username := defaultOptions.env["QUERY_DEFAULT_USER"]
	password := defaultOptions.env["QUERY_DEFAULT_PASSWORD"]
	if password == "" && username == "" {
		return nil, errors.New("empty password and user")
	}

	// module options take precedence over default options
	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env))

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *DatabendContainer
	if container != nil {
		c = &DatabendContainer{
			Container: container,
			password:  password,
			username:  username,
			database:  defaultDatabaseName,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// MustConnectionString panics if the address cannot be determined.
func (c *DatabendContainer) MustConnectionString(ctx context.Context, args ...string) string {
	addr, err := c.ConnectionString(ctx, args...)
	if err != nil {
		panic(err)
	}
	return addr
}

func (c *DatabendContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "8000/tcp", "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	extraArgs := ""
	if len(args) > 0 {
		extraArgs = "?" + strings.Join(args, "&")
	}
	if c.database == "" {
		return "", errors.New("database name is empty")
	}

	// databend://databend:databend@localhost:8000/default?sslmode=disable
	connectionString := fmt.Sprintf("databend://%s:%s@%s/%s%s", c.username, c.password, endpoint, c.database, extraArgs)
	return connectionString, nil
}
