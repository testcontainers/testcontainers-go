package ravendb

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultPort = "8080/tcp"

// Container represents the RavenDB container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the RavenDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithEnv(map[string]string{
			"RAVEN_Setup_Mode":                      "None",
			"RAVEN_License_Eula_Accepted":           "true",
			"RAVEN_Security_UnsecuredAccessAllowed": "PublicNetwork",
			"RAVEN_Logs_Mode":                       "None",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForHTTP("/health/server").
					WithPort(defaultPort).
					WithStatusCodeMatcher(func(status int) bool {
						return status == 200
					}),
			).WithDeadline(120*time.Second),
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
	endpoint, err := c.PortEndpoint(ctx, defaultPort, "http")
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return endpoint, nil
}
