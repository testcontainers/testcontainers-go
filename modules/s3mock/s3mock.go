// Package s3mock provides a testcontainers module for Adobe S3Mock, a lightweight
// server that implements the AWS S3 API. Use it in tests to interact with S3
// buckets and objects without real AWS credentials or network access.
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
// It wraps the generic testcontainers.Container and exposes S3Mock-specific
// helper methods for retrieving the HTTP and HTTPS endpoint URLs.
type Container struct {
	testcontainers.Container
}

// WithInitialBuckets returns an option that configures S3Mock to pre-create
// the given buckets on startup. Calling it with no arguments is a no-op.
// Both the 3.x/4.x env var (domain prefix) and the 5.x+ env var (store prefix)
// are set so that the option works across all supported S3Mock versions.
func WithInitialBuckets(buckets ...string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if len(buckets) == 0 {
			return nil
		}
		joined := strings.Join(buckets, ",")
		// 3.x / 4.x: com.adobe.testing.s3mock.domain.initialBuckets
		req.Env["COM_ADOBE_TESTING_S3MOCK_DOMAIN_INITIAL_BUCKETS"] = joined
		// 5.x+: com.adobe.testing.s3mock.store.initialBuckets
		req.Env["COM_ADOBE_TESTING_S3MOCK_STORE_INITIAL_BUCKETS"] = joined
		return nil
	}
}

// EndpointURL returns the HTTP endpoint URL for the S3Mock container,
// using the dynamically mapped host port for container port 9090.
func (c *Container) EndpointURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpPort, "http")
}

// HTTPSEndpointURL returns the HTTPS endpoint URL for the S3Mock container,
// using the dynamically mapped host port for container port 9191.
func (c *Container) HTTPSEndpointURL(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, httpsPort, "https")
}

// Run creates and starts an S3Mock container using the given image and options.
// The container exposes HTTP on port 9090 and HTTPS on port 9191, and waits
// until the /favicon.ico health path returns HTTP 200 before returning.
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
