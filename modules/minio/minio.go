package minio

import (
	"context"
	"errors"
	"fmt"

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
		testcontainers.WithCmd("server", "/data"),
		testcontainers.WithExposedPorts("9000/tcp"),
		testcontainers.WithEnv(map[string]string{
			"MINIO_ROOT_USER":     defaultUser,
			"MINIO_ROOT_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/minio/health/live").WithPort("9000")),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("minio option: %w", err)
			}
		}
	}

	username := defaultOptions.env["MINIO_ROOT_USER"]
	password := defaultOptions.env["MINIO_ROOT_PASSWORD"]
	if username == "" || password == "" {
		return nil, errors.New("username or password has not been set")
	}

	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *MinioContainer
	if ctr != nil {
		c = &MinioContainer{Container: ctr, Username: username, Password: password}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}
