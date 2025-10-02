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
	defaultUser         = "databend"
	defaultPassword     = "databend"
	defaultDatabaseName = "default"
	defaultPort         = "8000/tcp"
)

// DatabendContainer represents the Databend container type used in the module
type DatabendContainer struct {
	testcontainers.Container
	username string
	password string
	database string
}

// Deprecated: use testcontainers.ContainerCustomizer instead
var _ testcontainers.ContainerCustomizer = (*DatabendOption)(nil)

// Deprecated: use testcontainers.ContainerCustomizer instead
// DatabendOption is an option for the Databend container.
type DatabendOption func(*DatabendContainer)

// Deprecated: use testcontainers.ContainerCustomizer instead
// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o DatabendOption) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// Run creates an instance of the Databend container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DatabendContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"QUERY_DEFAULT_USER":     defaultUser,
			"QUERY_DEFAULT_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultPort)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *DatabendContainer
	if ctr != nil {
		// set default credentials
		c = &DatabendContainer{
			Container: ctr,
			password:  defaultPassword,
			username:  defaultUser,
			database:  defaultDatabaseName,
		}
	}

	if err != nil {
		return c, fmt.Errorf("run databend: %w", err)
	}

	// refresh the credentials from the environment variables
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect databend: %w", err)
	}

	foundUser, foundPass := false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "QUERY_DEFAULT_USER="); ok {
			c.username, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "QUERY_DEFAULT_PASSWORD="); ok {
			c.password, foundPass = v, true
		}

		if foundUser && foundPass {
			break
		}
	}

	if c.username == "" && c.password == "" {
		return c, errors.New("empty password and user")
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

// WithUsername sets the username for the Databend container.
// WithUsername is [Run] option that configures the default query user by setting
// the `QUERY_DEFAULT_USER` container environment variable.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["QUERY_DEFAULT_USER"] = username
		return nil
	}
}

// WithPassword sets the password for the Databend container.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["QUERY_DEFAULT_PASSWORD"] = password
		return nil
	}
}
