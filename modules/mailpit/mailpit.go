package mailpit

import (
	"context"
	"fmt"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	smtpPort = "1025/tcp"
	httpPort = "8025/tcp"
)

// Container represents the Mailpit container type used in the module
type Container struct {
	testcontainers.Container
}

// SMTPEndpoint returns the SMTP endpoint of the Mailpit container, in the
// form "host:port" using the 1025/tcp port.
func (c *Container) SMTPEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, smtpPort, "")
}

// HTTPURL returns the URL of the Mailpit HTTP interface and REST API,
// using the 8025/tcp port.
func (c *Container) HTTPURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// WithSMTPAuth configures SMTP authentication credentials.
func WithSMTPAuth(user, password string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MP_SMTP_AUTH_USERNAME": user,
		"MP_SMTP_AUTH_PASSWORD": password,
	})
}

// WithMessageLimit sets the maximum number of messages to keep in Mailpit.
// When the limit is reached the oldest messages are deleted.
func WithMessageLimit(limit int) testcontainers.CustomizeRequestOption {
	return testcontainers.WithEnv(map[string]string{
		"MP_MAX_MESSAGES": strconv.Itoa(limit),
	})
}

// Run creates an instance of the Mailpit container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(smtpPort, httpPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/api/v1/info").
				WithPort(httpPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
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
		return c, fmt.Errorf("run mailpit: %w", err)
	}

	return c, nil
}
