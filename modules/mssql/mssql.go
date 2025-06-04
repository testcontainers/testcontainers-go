package mssql

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
	defaultUsername = "sa" // default microsoft system administrator
	defaultPassword = "Strong@Passw0rd"
)

// MSSQLServerContainer represents the MSSQLServer container type used in the module
type MSSQLServerContainer struct {
	testcontainers.Container
	password string
	username string
}

// Password returns the password for the MSSQLServer container
func (c *MSSQLServerContainer) Password() string {
	return c.password
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MSSQLServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error) {
	return Run(ctx, "mcr.microsoft.com/mssql/server:2022-CU18-ubuntu-22.04", opts...)
}

// Run creates an instance of the MSSQLServer container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort(defaultPort).WithStartupTimeout(time.Minute),
			wait.ForLog("Recovery is complete."),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultSettings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultSettings); err != nil {
				return nil, fmt.Errorf("mssql option: %w", err)
			}
		}
	}

	if strings.ToUpper(defaultSettings.env["ACCEPT_EULA"]) != "Y" {
		return nil, errors.New("EULA not accepted. Please use the WithAcceptEULA option to accept the EULA")
	}

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultSettings.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MSSQLServerContainer
	if ctr != nil {
		c = &MSSQLServerContainer{Container: ctr, password: defaultSettings.env["MSSQL_SA_PASSWORD"], username: defaultUsername}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// ConnectionString returns the connection string for the MSSQLServer container
func (c *MSSQLServerContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	containerPort, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	extraArgs := strings.Join(args, "&")

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?%s", c.username, c.password, host, containerPort.Port(), extraArgs)

	return connStr, nil
}
