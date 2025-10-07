package mssql

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
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

// WithAcceptEULA sets the ACCEPT_EULA environment variable to "Y"
func WithAcceptEULA() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ACCEPT_EULA"] = "Y"

		return nil
	}
}

// WithPassword sets the MSSQL_SA_PASSWORD environment variable to the provided password
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if password == "" {
			password = defaultPassword
		}
		req.Env["MSSQL_SA_PASSWORD"] = password

		return nil
	}
}

// WithInitSQL adds SQL scripts to be executed after the container is ready.
// The scripts are executed in the order they are provided using sqlcmd tool.
func WithInitSQL(files ...io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		hooks := make([]testcontainers.ContainerHook, 0, len(files))

		for i, script := range files {
			content, err := io.ReadAll(script)
			if err != nil {
				return fmt.Errorf("failed to read script: %w", err)
			}

			hook := func(ctx context.Context, c testcontainers.Container) error {
				password := defaultPassword
				if req.Env["MSSQL_SA_PASSWORD"] != "" {
					password = req.Env["MSSQL_SA_PASSWORD"]
				}

				// targetPath is a dummy path to store the script in the container
				targetPath := "/tmp/" + fmt.Sprintf("script_%d.sql", i)
				if err := c.CopyToContainer(ctx, content, targetPath, 0o644); err != nil {
					return fmt.Errorf("failed to copy script to container: %w", err)
				}

				// NOTE: we add both legacy and new mssql-tools paths to ensure compatibility
				envOpts := tcexec.WithEnv([]string{
					"PATH=/opt/mssql-tools18/bin:/opt/mssql-tools/bin:$PATH",
				})
				cmd := []string{
					"sqlcmd",
					"-S", "localhost",
					"-U", defaultUsername,
					"-P", password,
					"-No",
					"-i", targetPath,
				}
				if _, _, err := c.Exec(ctx, cmd, envOpts); err != nil {
					return fmt.Errorf("failed to execute SQL script %q using sqlcmd: %w", targetPath, err)
				}
				return nil
			}
			hooks = append(hooks, hook)
		}

		req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{
			PostReadies: hooks,
		})

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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"MSSQL_SA_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort(defaultPort).WithStartupTimeout(time.Minute),
			wait.ForLog("Recovery is complete."),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Validate EULA acceptance after applying user options
	validateEULA := func(req *testcontainers.GenericContainerRequest) error {
		if strings.ToUpper(req.Env["ACCEPT_EULA"]) != "Y" {
			return errors.New("EULA not accepted. Please use the WithAcceptEULA option to accept the EULA")
		}
		return nil
	}

	moduleOpts = append(moduleOpts, testcontainers.CustomizeRequestOption(validateEULA))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MSSQLServerContainer
	if ctr != nil {
		c = &MSSQLServerContainer{Container: ctr, username: defaultUsername}
	}

	if err != nil {
		return c, fmt.Errorf("run mssql: %w", err)
	}

	// Retrieve password from container environment
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect mssql: %w", err)
	}

	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "MSSQL_SA_PASSWORD="); ok {
			c.password = v
			break
		}
	}

	return c, nil
}

// ConnectionString returns the connection string for the MSSQLServer container
func (c *MSSQLServerContainer) ConnectionString(ctx context.Context, args ...string) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	extraArgs := strings.Join(args, "&")

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s?%s", c.username, c.password, endpoint, extraArgs)

	return connStr, nil
}
