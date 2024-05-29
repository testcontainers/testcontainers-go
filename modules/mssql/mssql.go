package mssql

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultImage    = "mcr.microsoft.com/mssql/server:2022-CU10-ubuntu-22.04"
	defaultPort     = "1433/tcp"
	defaultUsername = "sa" // default microsoft system administrator
	defaultPassword = "Strong@Passw0rd"
)

// MSSQLServerContainer represents the MSSQLServer container type used in the module
type MSSQLServerContainer struct {
	*testcontainers.DockerContainer
	password string
	username string
}

func WithAcceptEULA() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
		req.Env["ACCEPT_EULA"] = "Y"

		return nil
	}
}

func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.Request) error {
		if password == "" {
			password = defaultPassword
		}
		req.Env["MSSQL_SA_PASSWORD"] = password

		return nil
	}
}

// RunContainer creates an instance of the MSSQLServer container type
func RunContainer(ctx context.Context, opts ...testcontainers.RequestCustomizer) (*MSSQLServerContainer, error) {
	req := testcontainers.Request{
		Image:        defaultImage,
		ExposedPorts: []string{defaultPort},
		Env: map[string]string{
			"MSSQL_SA_PASSWORD": defaultPassword,
		},
		WaitingFor: wait.ForLog("Recovery is complete."),
		Started:    true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.New(ctx, req)
	if err != nil {
		return nil, err
	}

	username := defaultUsername
	password := req.Env["MSSQL_SA_PASSWORD"]

	return &MSSQLServerContainer{DockerContainer: container, password: password, username: username}, nil
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
