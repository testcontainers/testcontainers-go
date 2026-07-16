package sqledge

import (
	"context"
	"errors"
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

// WithAcceptEULA sets the ACCEPT_EULA environment variable to "Y", indicating
// acceptance of the Microsoft Azure SQL Edge End-User License Agreement.
// This option is required; Run returns an error if it is not provided.
func WithAcceptEULA() testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"ACCEPT_EULA": "Y",
	})
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

// Run creates an instance of the Azure SQL Edge container type.
// Callers must pass WithAcceptEULA() to accept the Microsoft license agreement.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"MSSQL_SA_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(defaultPort).WithStartupTimeout(2*time.Minute),
			wait.ForLog("Recovery is complete."),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	// Validate EULA acceptance after applying user options.
	validateEULA := func(req *testcontainers.GenericContainerRequest) error {
		if strings.ToUpper(req.Env["ACCEPT_EULA"]) != "Y" {
			return errors.New("EULA not accepted: use WithAcceptEULA() to accept the Azure SQL Edge license agreement")
		}
		return nil
	}
	moduleOpts = append(moduleOpts, testcontainers.CustomizeRequestOption(validateEULA))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run sqledge: %w", err)
	}

	// Retrieve the effective password from container environment.
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
// using the default 1433 port. Additional query parameters may be appended via args.
// Example: container.ConnectionString(ctx, "encrypt=false", "TrustServerCertificate=true")
func (c *Container) ConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=master", defaultUsername, c.password, host, port.Port())
	if len(args) > 0 {
		connStr += "&" + strings.Join(args, "&")
	}

	return connStr, nil
}
