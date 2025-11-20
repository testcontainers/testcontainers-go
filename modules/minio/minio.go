package minio

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultUser     = "minioadmin"
	defaultPassword = "minioadmin"
)

// MinioContainer represents the Minio container type used in the module
type MinioContainer struct {
	testcontainers.Container
	Username string
	Password string
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user. It must not be empty or undefined.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MINIO_ROOT_USER"] = username

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the Minio image. It must not be empty or undefined.
// This environment variable sets the root user password for Minio.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MINIO_ROOT_PASSWORD"] = password

		return nil
	}
}

// ConnectionString returns the connection string for the minio container, using the default 9000 port, and
// obtaining the host and exposed port from the container.
func (c *MinioContainer) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, "9000/tcp", "")
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Minio container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error) {
	return Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z", opts...)
}

// Run creates an instance of the Minio container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("9000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/minio/health/live").WithPort("9000")),
		testcontainers.WithEnv(map[string]string{
			"MINIO_ROOT_USER":     defaultUser,
			"MINIO_ROOT_PASSWORD": defaultPassword,
		}),
		testcontainers.WithCmd("server", "/data"),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Validate credentials after applying user options
	validateCreds := func(req *testcontainers.GenericContainerRequest) error {
		username := req.Env["MINIO_ROOT_USER"]
		password := req.Env["MINIO_ROOT_PASSWORD"]
		if username == "" || password == "" {
			return errors.New("username or password has not been set")
		}
		return nil
	}

	moduleOpts = append(moduleOpts, testcontainers.CustomizeRequestOption(validateCreds))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MinioContainer
	if ctr != nil {
		c = &MinioContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run minio: %w", err)
	}

	// Retrieve credentials from container environment
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect minio: %w", err)
	}

	var foundUser, foundPass bool
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "MINIO_ROOT_USER="); ok {
			c.Username, foundUser = v, true
		}
		if v, ok := strings.CutPrefix(env, "MINIO_ROOT_PASSWORD="); ok {
			c.Password, foundPass = v, true
		}

		if foundUser && foundPass {
			break
		}
	}

	return c, nil
}
