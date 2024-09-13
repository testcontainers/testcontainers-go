package gcloud

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
)

const defaultProjectID = "test-project"

type Container struct {
	*testcontainers.DockerContainer
	Settings options
	URI      string
}

// newGCloudContainer creates a new GCloud container, obtaining the URL to access the container from the specified port.
func newGCloudContainer(ctx context.Context, req testcontainers.Request, port int, settings options, urlPrefix string) (*Container, error) {
	ctr, err := testcontainers.Run(ctx, req)
	var c *Container
	if ctr != nil {
		c = &Container{DockerContainer: ctr, Settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", port)))
	if err != nil {
		return c, fmt.Errorf("mapped port: %w", err)
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return c, fmt.Errorf("host: %w", err)
	}

	c.URI = urlPrefix + hostIP + ":" + mappedPort.Port()

	return c, nil
}

type options struct {
	ProjectID string
}

func defaultOptions() options {
	return options{
		ProjectID: defaultProjectID,
	}
}

// Compiler check to ensure that Option implements the testcontainers.RequestCustomizer interface.
var _ testcontainers.RequestCustomizer = (*Option)(nil)

// Option is an option for the GCloud container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.RequestCustomizer interface.
func (o Option) Customize(*testcontainers.Request) error {
	// NOOP to satisfy interface.
	return nil
}

// WithProjectID sets the project ID for the GCloud container.
func WithProjectID(projectID string) Option {
	return func(o *options) {
		o.ProjectID = projectID
	}
}

// applyOptions applies the options to the container request and returns the settings.
func applyOptions(req *testcontainers.Request, opts []testcontainers.RequestCustomizer) (options, error) {
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		if err := opt.Customize(req); err != nil {
			return options{}, err
		}
	}

	return settings, nil
}
