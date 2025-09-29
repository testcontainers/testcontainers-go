package artemis

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultBrokerPort = "61616/tcp"
	defaultHTTPPort   = "8161/tcp"
	defaultUser       = "artemis"
	defaultPassword   = "artemis"
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

// WithCredentials sets the administrator credentials. The default is artemis:artemis.
func WithCredentials(user, password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ARTEMIS_USER"] = user
		req.Env["ARTEMIS_PASSWORD"] = password

		return nil
	}
}

// WithAnonymousLogin enables anonymous logins.
func WithAnonymousLogin() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ANONYMOUS_LOGIN"] = "true"

		return nil
	}
}

// Additional arguments sent to the `artemis create` command.
// The default is `--http-host 0.0.0.0 --relax-jolokia`.
// Setting this value will override the default.
// See the documentation on `artemis create` for available options.
func WithExtraArgs(args string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["EXTRA_ARGS"] = args

		return nil
	}
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
		testcontainers.WithEnv(map[string]string{
			"ARTEMIS_USER":     defaultUser,
			"ARTEMIS_PASSWORD": defaultPassword,
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort(defaultBrokerPort),
			wait.ForListeningPort(defaultHTTPPort),
			wait.ForLog("Server is now live"),
			wait.ForLog("REST API available"),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}
	if err != nil {
		return c, fmt.Errorf("run artemis: %w", err)
	}

	// initialize the credentials
	c.user = defaultUser
	c.password = defaultPassword

	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect artemis: %w", err)
	}

	// refresh the credentials from the environment variables
	foundUser, foundPass := false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "ARTEMIS_USER="); ok {
			c.user, foundUser = v, true
			continue
		} else if v, ok := strings.CutPrefix(env, "ARTEMIS_PASSWORD="); ok {
			c.password, foundPass = v, true
			continue
		}

		if foundUser && foundPass {
			break
		}
	}

	return c, nil
}
