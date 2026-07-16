package s3mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	httpPort  = "9090/tcp"
	httpsPort = "9191/tcp"
)

// Container represents the S3Mock container type used in the module.
type Container struct {
	testcontainers.Container
}

// WithInitialBuckets configures S3Mock to pre-create the given buckets on startup.
func WithInitialBuckets(buckets ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if len(buckets) == 0 {
			return nil
		}
		if req.Env == nil {
			req.Env = map[string]string{}
		}
		req.Env["COM_ADOBE_TESTING_S3MOCK_STORE_INITIAL_BUCKETS"] = strings.Join(buckets, ",")
		return nil
	}
}

// EndpointURL returns the HTTP endpoint URL for the S3Mock container (mapped from container port 9090).
func (c *Container) EndpointURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// HTTPSEndpointURL returns the HTTPS endpoint URL for the S3Mock container (mapped from container port 9191).
func (c *Container) HTTPSEndpointURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpsPort, "https")
}

// Run creates an instance of the S3Mock container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 3+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(httpPort, httpsPort),
		testcontainers.WithEnv(map[string]string{}),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/favicon.ico").WithPort(httpPort),
		),
	)

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run s3mock: %w", err)
	}

	return c, nil
}
