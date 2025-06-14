package artemis

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultBrokerPort = "61616/tcp"
	defaultHTTPPort   = "8161/tcp"
)

// Container represents the Artemis container type used in the module.
type Container struct {
	testcontainers.Container
	user     string
	password string
}

// User returns the administrator username.
func (c *Container) User() string {
	return c.user
}

// Password returns the administrator password.
func (c *Container) Password() string {
	return c.password
}

// BrokerEndpoint returns the host:port for the combined protocols endpoint.
// The endpoint accepts CORE, MQTT, AMQP, STOMP, HORNETQ and OPENWIRE protocols.
func (c *Container) BrokerEndpoint(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, nat.Port(defaultBrokerPort), "")
}

// ConsoleURL returns the URL for the management console.
func (c *Container) ConsoleURL(ctx context.Context) (string, error) {
	host, err := c.PortEndpoint(ctx, nat.Port(defaultHTTPPort), "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%s@%s/console", c.user, c.password, host), nil
}

// Deprecated: use Run instead.
// RunContainer creates an instance of the Artemis container type.
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	return Run(ctx, "apache/activemq-artemis:2.30.0-alpine", opts...)
}

// Run creates an instance of the Artemis container type with a given image
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultBrokerPort, defaultHTTPPort),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("Server is now live"),
			wait.ForLog("REST API available"),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("artemis option: %w", err)
			}
		}
	}

	// module options take precedence over default options
	moduleOpts = append(moduleOpts, testcontainers.WithEnv(defaultOptions.env))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, user: defaultOptions.env["ARTEMIS_USER"], password: defaultOptions.env["ARTEMIS_PASSWORD"]}
	}
	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}
