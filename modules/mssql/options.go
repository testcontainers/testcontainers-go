package mssql

import (
	"context"
	"fmt"
	"io"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"MSSQL_SA_PASSWORD": defaultPassword,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the MSSQL container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAcceptEULA sets the ACCEPT_EULA environment variable to "Y"
func WithAcceptEULA() Option {
	return func(o *options) error {
		o.env["ACCEPT_EULA"] = "Y"

		return nil
	}
}

// WithPassword sets the MSSQL_SA_PASSWORD environment variable to the provided password
func WithPassword(password string) Option {
	return func(o *options) error {
		if password == "" {
			password = defaultPassword
		}
		o.env["MSSQL_SA_PASSWORD"] = password

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
				envOpts := exec.WithEnv([]string{
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
