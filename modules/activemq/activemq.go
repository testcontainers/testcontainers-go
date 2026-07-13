package activemq

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultBrokerPort     = "61616/tcp"
	defaultWebConsolePort = "8161/tcp"
	defaultAdminUser      = "admin"
	defaultAdminPassword  = "admin"
)

// Container represents the ActiveMQ container type used in the module.
type Container struct {
	testcontainers.Container
	adminUser     string
	adminPassword string
}

// AdminUser returns the administrator username for the ActiveMQ web console.
func (c *Container) AdminUser() string {
	return c.adminUser
}

// AdminPassword returns the administrator password for the ActiveMQ web console.
func (c *Container) AdminPassword() string {
	return c.adminPassword
}

// BrokerURL returns the OpenWire broker URL for the mapped 61616/tcp port.
func (c *Container) BrokerURL(ctx context.Context) (string, error) {
	hostPort, err := c.PortEndpoint(ctx, defaultBrokerPort, "")
	if err != nil {
		return "", err
	}
	return "tcp://" + hostPort, nil
}

// WebConsoleURL returns the HTTP URL of the ActiveMQ web console for the mapped 8161/tcp port.
func (c *Container) WebConsoleURL(ctx context.Context) (string, error) {
	hostPort, err := c.PortEndpoint(ctx, defaultWebConsolePort, "")
	if err != nil {
		return "", err
	}
	return "http://" + hostPort, nil
}

// WithAdminCredentials sets the username and password for the ActiveMQ web console.
// The credentials are passed via the ACTIVEMQ_WEB_ADMIN_NAME and
// ACTIVEMQ_WEB_ADMIN_PASSWORD environment variables.
// Note: the Jolokia REST API always requires the built-in admin/admin credentials
// regardless of these environment variables, so the wait strategy is not overridden.
func WithAdminCredentials(user, password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["ACTIVEMQ_WEB_ADMIN_NAME"] = user
		req.Env["ACTIVEMQ_WEB_ADMIN_PASSWORD"] = password
		return nil
	}
}

// Run creates an instance of the ActiveMQ container type with a given image.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultBrokerPort, defaultWebConsolePort),
		testcontainers.WithEnv(map[string]string{
			"ACTIVEMQ_BROKER_NAME":        "localhost",
			"ACTIVEMQ_WEB_ADMIN_NAME":     defaultAdminUser,
			"ACTIVEMQ_WEB_ADMIN_PASSWORD": defaultAdminPassword,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(defaultBrokerPort),
			wait.ForHTTP("/api/jolokia/version").
				WithPort(defaultWebConsolePort).
				WithBasicAuth(defaultAdminUser, defaultAdminPassword).
				WithStatusCodeMatcher(func(status int) bool { return status == http.StatusOK }).
				WithStartupTimeout(60*time.Second),
		),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{
			Container:     ctr,
			adminUser:     defaultAdminUser,
			adminPassword: defaultAdminPassword,
		}
	}
	if err != nil {
		return c, fmt.Errorf("run activemq: %w", err)
	}

	// Refresh credentials from the running container's environment.
	inspect, err := ctr.Inspect(ctx)
	if err != nil {
		return c, fmt.Errorf("inspect activemq: %w", err)
	}
	foundUser, foundPass := false, false
	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "ACTIVEMQ_WEB_ADMIN_NAME="); ok {
			c.adminUser, foundUser = v, true
		} else if v, ok := strings.CutPrefix(env, "ACTIVEMQ_WEB_ADMIN_PASSWORD="); ok {
			c.adminPassword, foundPass = v, true
		}
		if foundUser && foundPass {
			break
		}
	}

	return c, nil
}
