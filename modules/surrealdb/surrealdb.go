package surrealdb

import (
	"context"
	"fmt"
	"net"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SurrealDBContainer represents the SurrealDB container type used in the module
type SurrealDBContainer struct {
	testcontainers.Container
}

// ConnectionString returns the connection string for the OpenLDAP container
func (c *SurrealDBContainer) URL(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "8000/tcp")
	if err != nil {
		return "", err
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	connStr := fmt.Sprintf("ws://%s/rpc", net.JoinHostPort(host, containerPort.Port()))
	return connStr, nil
}

// WithUser sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["SURREAL_USER"] = username
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for SurrealDB.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["SURREAL_PASS"] = password
	}
}

// WithAuthentication enables authentication for the SurrealDB instance
func WithAuthentication() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["SURREAL_AUTH"] = "true"
	}
}

// WithStrict enables strict mode for the SurrealDB instance
func WithStrictMode() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["SURREAL_STRICT"] = "true"
	}
}

// WithAllowAllCaps enables all caps for the SurrealDB instance
func WithAllowAllCaps() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) {
		req.Env["SURREAL_CAPS_ALLOW_ALL"] = "false"
	}
}

// RunContainer creates an instance of the SurrealDB container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*SurrealDBContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "surrealdb/surrealdb:v1.1.1",
		Env: map[string]string{
			"SURREAL_USER":           "root",
			"SURREAL_PASS":           "root",
			"SURREAL_AUTH":           "false",
			"SURREAL_STRICT":         "false",
			"SURREAL_CAPS_ALLOW_ALL": "false",
			"SURREAL_PATH":           "memory",
		},
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Started web server on "),
		),
		Cmd: []string{"start"},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &SurrealDBContainer{Container: container}, nil
}
