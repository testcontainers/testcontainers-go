package surrealdb

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SurrealDBContainer represents the SurrealDB container type used in the module
type SurrealDBContainer struct {
	testcontainers.Container
}

// URL returns the connection string for the OpenLDAP container
func (c *SurrealDBContainer) URL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, "8000/tcp", "ws")
	if err != nil {
		return "", err
	}

	return endpoint + "/rpc", nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{"SURREAL_USER": username})(req)
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for SurrealDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{"SURREAL_PASS": password})(req)
	}
}

// WithAuthentication enables authentication for the SurrealDB instance
func WithAuthentication() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{"SURREAL_AUTH": "true"})(req)
	}
}

// WithStrictMode enables strict mode for the SurrealDB instance
func WithStrictMode() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{"SURREAL_STRICT": "true"})(req)
	}
}

// WithAllowAllCaps enables all caps for the SurrealDB instance
func WithAllowAllCaps() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		return testcontainers.WithEnv(map[string]string{"SURREAL_CAPS_ALLOW_ALL": "false"})(req)
	}
}

// Deprecated: use Run instead
// RunContainer creates an instance of the SurrealDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SurrealDBContainer, error) {
	return Run(ctx, "surrealdb/surrealdb:v1.1.1", opts...)
}

// Run creates an instance of the SurrealDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*SurrealDBContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithEnv(map[string]string{
			"SURREAL_USER":           "root",
			"SURREAL_PASS":           "root",
			"SURREAL_AUTH":           "false",
			"SURREAL_STRICT":         "false",
			"SURREAL_CAPS_ALLOW_ALL": "false",
			"SURREAL_PATH":           "memory",
		}),
		testcontainers.WithExposedPorts("8000/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Started web server on ")),
		testcontainers.WithCmd("start"),
	}

	ctr, err := testcontainers.Run(ctx, img, append(moduleOpts, opts...)...)
	var c *SurrealDBContainer
	if ctr != nil {
		c = &SurrealDBContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run surrealdb: %w", err)
	}

	return c, nil
}
