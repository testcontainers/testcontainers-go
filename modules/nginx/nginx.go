package nginx

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// httpPort is the default HTTP port for Nginx.
	httpPort = "80/tcp"
	// httpsPort is the default HTTPS port for Nginx.
	httpsPort = "443/tcp"

	// nginxConfPath is the path to the main nginx configuration file.
	nginxConfPath = "/etc/nginx/nginx.conf"
	// nginxDefaultConfPath is the path to the default site configuration snippet.
	nginxDefaultConfPath = "/etc/nginx/conf.d/default.conf"
)

// Container represents the Nginx container type used in the module.
type Container struct {
	testcontainers.Container
}

// WithConfigFile mounts a custom nginx.conf file at /etc/nginx/nginx.conf in the container.
// The hostPath must be an absolute path to the configuration file on the host.
func WithConfigFile(hostPath string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      hostPath,
		ContainerFilePath: nginxConfPath,
		FileMode:          0o644,
	})
}

// WithCustomConfig mounts a configuration snippet at /etc/nginx/conf.d/default.conf in the container.
// The hostPath must be an absolute path to the configuration file on the host.
func WithCustomConfig(hostPath string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithFiles(testcontainers.ContainerFile{
		HostFilePath:      hostPath,
		ContainerFilePath: nginxDefaultConfPath,
		FileMode:          0o644,
	})
}

// Run creates an instance of the Nginx container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort, httpsPort),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(httpPort)),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run nginx: %w", err)
	}

	return c, nil
}

// HTTPEndpoint returns the HTTP endpoint of the Nginx container, in the form "http://host:port".
func (c *Container) HTTPEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, httpPort, "http")
	if err != nil {
		return "", fmt.Errorf("http endpoint: %w", err)
	}

	return endpoint, nil
}

// HTTPSEndpoint returns the HTTPS endpoint of the Nginx container, in the form "https://host:port".
func (c *Container) HTTPSEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, httpsPort, "https")
	if err != nil {
		return "", fmt.Errorf("https endpoint: %w", err)
	}

	return endpoint, nil
}
