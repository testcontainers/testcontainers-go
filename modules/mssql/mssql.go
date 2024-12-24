package mssql

import (
	"context"
	"fmt"
	"strings"

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

func WithAcceptEULA() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ACCEPT_EULA"] = "Y"

		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password == "" {
			password = defaultPassword
		}
		req.Env["MSSQL_SA_PASSWORD"] = password

		return nil
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the MSSQLServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error) {
	return Run(ctx, "mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04", opts...)
}

// Run creates an instance of the MSSQLServer container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MSSQLServerContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{defaultPort},
		Env: map[string]string{
			"MSSQL_SA_PASSWORD": defaultPassword,
		},
		WaitingFor: wait.ForLog("Recovery is complete."),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MSSQLServerContainer
	if container != nil {
		c = &MSSQLServerContainer{Container: container, password: req.Env["MSSQL_SA_PASSWORD"], username: defaultUsername}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}

func (c *MSSQLServerContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	containerPort, err := c.MappedPort(ctx, defaultPort)
	if err != nil {
		return "", err
	}

	extraArgs := strings.Join(args, "&")

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%s?%s", c.username, c.password, host, containerPort.Port(), extraArgs)

	return connStr, nil
}
