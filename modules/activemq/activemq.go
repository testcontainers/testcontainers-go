// Package activemq provides a testcontainers module for Apache ActiveMQ Classic.
package activemq

import (
	"context"
	"fmt"

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

// AdminUser returns the web-console username.
// The ActiveMQ Classic image protects the web console with Jetty's
// HashLoginService (conf/jetty-realm.properties), which is hardcoded to
// "admin: admin, admin" and cannot be overridden via environment variables.
// This method therefore always returns "admin".
func (c *Container) AdminUser() string {
	return c.adminUser
}

// AdminPassword returns the web-console password.
// The ActiveMQ Classic image protects the web console with Jetty's
// HashLoginService (conf/jetty-realm.properties), which is hardcoded to
// "admin: admin, admin" and cannot be overridden via environment variables.
// This method therefore always returns "admin".
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

// WithAdminCredentials sets ACTIVEMQ_WEB_USER and ACTIVEMQ_WEB_PASSWORD.
// The image entrypoint uses these to update conf/users.properties, which
// configures the JAAS PropertiesLoginModule used for broker connection/messaging
// security when connection-level authentication is enabled.
//
// NOTE: these variables do NOT affect web-console or Jolokia authentication.
// The web console and /api/jolokia/* endpoints are protected by Jetty's
// HashLoginService reading conf/jetty-realm.properties, which is hardcoded to
// "admin: admin, admin" and is not configurable via environment variables.
// AdminUser() and AdminPassword() always return those built-in values.
func WithAdminCredentials(user, password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["ACTIVEMQ_WEB_USER"] = user
		req.Env["ACTIVEMQ_WEB_PASSWORD"] = password
		return nil
	}
}

// Run creates an instance of the ActiveMQ container type with a given image.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultBrokerPort, defaultWebConsolePort),
		testcontainers.WithEnv(map[string]string{
			"ACTIVEMQ_BROKER_NAME":  "localhost",
			"ACTIVEMQ_WEB_USER":     defaultAdminUser,
			"ACTIVEMQ_WEB_PASSWORD": defaultAdminPassword,
		}),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog(".*Apache ActiveMQ.*started.*").AsRegexp(),
				wait.ForListeningPort(defaultBrokerPort),
			),
		),
	)
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		// adminUser and adminPassword are always the Jetty-realm values; they
		// cannot be changed via environment variables in this image.
		c = &Container{
			Container:     ctr,
			adminUser:     defaultAdminUser,
			adminPassword: defaultAdminPassword,
		}
	}
	if err != nil {
		return c, fmt.Errorf("run activemq: %w", err)
	}

	return c, nil
}
