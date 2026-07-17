// Package papercutsmtp provides a testcontainers-go module for Papercut SMTP,
// a lightweight SMTP server with a built-in web UI for capturing and inspecting
// outbound emails in tests. No messages are relayed to real recipients.
package papercutsmtp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// smtpPort is the default SMTP port exposed by the Papercut SMTP container.
	smtpPort = "2525/tcp"

	// httpPort is the default HTTP port for the Papercut SMTP web UI.
	httpPort = "8080/tcp"
)

// Container represents the PapercutSMTP container type used in the module
type Container struct {
	testcontainers.Container
}

// SMTPEndpoint returns the host:port endpoint for the SMTP server (port 2525).
func (c *Container) SMTPEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, smtpPort, "")
}

// HTTPURL returns the URL for the Papercut SMTP web UI (port 8080).
func (c *Container) HTTPURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// Run creates an instance of the PapercutSMTP container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(smtpPort, httpPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/").WithPort(httpPort).WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
		),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run papercutsmtp: %w", err)
	}

	return c, nil
}
