package sftp

import (
	"context"
	"errors"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultSSHPort = "22/tcp"

// Container represents the SFTP container type used in the module.
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the SFTP container type.
// At least one user must be configured via WithUser, otherwise an error is returned.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Gather all config options (defaults and then apply provided options).
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(settings)
		}
	}

	if len(settings.users) == 0 {
		return nil, errors.New("run sftp: at least one user is required")
	}

	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultSSHPort),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(defaultSSHPort)),
		testcontainers.WithCmd(settings.users...),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run sftp: %w", err)
	}

	return c, nil
}

// Address returns the host:port address of the SFTP server, suitable for use
// with an SSH or SFTP client dial call.
func (c *Container) Address(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("sftp host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultSSHPort)
	if err != nil {
		return "", fmt.Errorf("sftp port: %w", err)
	}

	return fmt.Sprintf("%s:%s", host, port.Port()), nil
}
