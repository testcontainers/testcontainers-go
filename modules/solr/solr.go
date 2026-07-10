package solr

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultHTTPPort = "8983/tcp"
)

// Container represents the Solr container type used in the module.
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Solr container type.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := make([]testcontainers.ContainerCustomizer, 0, 2+len(opts))
	moduleOpts = append(moduleOpts,
		testcontainers.WithExposedPorts(defaultHTTPPort),
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/solr/admin/info/system?wt=json").
				WithPort(defaultHTTPPort).
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
		return c, fmt.Errorf("run solr: %w", err)
	}

	return c, nil
}

// WithCollection returns a [testcontainers.CustomizeRequestOption] that creates
// a named Solr collection after the container is ready.
func WithCollection(name string) testcontainers.CustomizeRequestOption {
	return testcontainers.WithAfterReadyCommand(
		testcontainers.NewRawCommand([]string{"solr", "create", "-c", name}),
	)
}

// Address returns the HTTP address of the Solr container in the form
// "http://<host>:<port>/solr".
func (c *Container) Address(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, defaultHTTPPort)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s/solr", host, port.Port()), nil
}

// CollectionURL returns the HTTP URL of a specific Solr collection in the form
// "http://<host>:<port>/solr/<collection>".
func (c *Container) CollectionURL(ctx context.Context, collection string) (string, error) {
	addr, err := c.Address(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", addr, collection), nil
}
