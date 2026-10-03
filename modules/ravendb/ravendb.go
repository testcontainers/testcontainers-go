package ravendb

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// defaultHTTPPort is the port used for the RavenDB HTTP/REST API and Studio UI.
	defaultHTTPPort = "8080/tcp"

	// defaultTCPPort is the port used for RavenDB subscriptions and cluster/replication communication.
	defaultTCPPort = "38888/tcp"
)

// Container represents the RavenDB container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the RavenDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultHTTPPort, defaultTCPPort),
		testcontainers.WithEnv(map[string]string{
			"RAVEN_Setup_Mode":                      "None",
			"RAVEN_License_Eula_Accepted":           "true",
			"RAVEN_Security_UnsecuredAccessAllowed": "PublicNetwork",
			"RAVEN_Logs_Mode":                       "None",
		}),
		testcontainers.WithWaitStrategyAndDeadline(
			120*time.Second,
			wait.ForHTTP("/build/version").
				WithPort(defaultHTTPPort).
				WithStatusCodeMatcher(func(status int) bool {
					return status == 200
				}).
				WithStartupTimeout(120*time.Second),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run ravendb: %w", err)
	}

	return c, nil
}

// ManagementURL returns the URL of the RavenDB management interface (Studio UI and REST API).
// The URL has the format http://<host>:<port>.
func (c *Container) ManagementURL(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultHTTPPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return endpoint, nil
}

// TCPEndpoint returns the host:port endpoint for RavenDB's TCP port (38888),
// used for subscriptions and cluster/replication communication.
func (c *Container) TCPEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultTCPPort, "")
	if err != nil {
		return "", fmt.Errorf("tcp port endpoint: %w", err)
	}

	return endpoint, nil
}
