package sqledge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort     = "1433/tcp"
	defaultUsername = "sa"
	defaultPassword = "Strong!Passw0rd"
)

// Container represents the Azure SQL Edge container type used in the module
type Container struct {
	testcontainers.Container
	password string
}

// WithPassword sets the MSSQL_SA_PASSWORD environment variable to the provided password.
// The password must meet SQL Server complexity requirements: uppercase + lowercase + number + special char.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password == "" {
			password = defaultPassword
		}

		return testcontainers.WithEnv(map[string]string{
			"MSSQL_SA_PASSWORD": password,
		})(req)
	}
}

// Run creates an instance of the Azure SQL Edge container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"ACCEPT_EULA":       "Y",
			"MSSQL_SA_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(defaultPort).WithStartupTimeout(2*time.Minute),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run sqledge: %w", err)
	}

	// Retrieve the effective password from container environment
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect sqledge: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "MSSQL_SA_PASSWORD="); ok {
			c.password = v
			break
		}
	}

	return c, nil
}

// ConnectionString returns the connection string for the Azure SQL Edge container,
// using the default 1433 port. TLS is disabled by default because azure-sql-edge
// ships a self-signed certificate with a negative serial number that Go 1.23+
// rejects at parse time. Additional query parameters may be appended via args
// and take precedence over the defaults.
// Example: container.ConnectionString(ctx, "database=mydb")
func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	// Disable TLS by default: azure-sql-edge uses a self-signed certificate
	// with a negative serial number, rejected by Go 1.23+ x509 validation.
	params := []string{"encrypt=disable", "TrustServerCertificate=true"}
	params = append(params, args...)
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=master&%s",
		defaultUsername, c.password, host, port.Port(), strings.Join(params, "&"))

	return connStr, nil
}
