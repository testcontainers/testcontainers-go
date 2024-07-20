package minio

import (
	"context"
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
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := c.MappedPort(ctx, "9000/tcp")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Minio container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error) {
	return Run(ctx, "docker.io/minio/minio:RELEASE.2024-01-16T16-07-38Z", opts...)
}

// Run creates an instance of the Minio container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*MinioContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForHTTP("/minio/health/live").WithPort("9000"),
		Env: map[string]string{
			"MINIO_ROOT_USER":     defaultUser,
			"MINIO_ROOT_PASSWORD": defaultPassword,
		},
		Cmd: []string{"server", "/data"},
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

	username := req.Env["MINIO_ROOT_USER"]
	password := req.Env["MINIO_ROOT_PASSWORD"]
	if username == "" || password == "" {
		return nil, fmt.Errorf("username or password has not been set")
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *MinioContainer
	if container != nil {
		c = &MinioContainer{Container: container, Username: username, Password: password}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
