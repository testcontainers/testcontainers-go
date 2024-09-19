package gcloud

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
)

const defaultProjectID = "test-project"

type GCloudContainer struct {
	testcontainers.Container
	Settings options
	URI      string
}

// newGCloudContainer creates a new GCloud container, obtaining the URL to access the container from the specified port.
func newGCloudContainer(ctx context.Context, req testcontainers.GenericContainerRequest, port int, settings options, urlPrefix string) (*GCloudContainer, error) {
	container, err := testcontainers.GenericContainer(ctx, req)
	var c *GCloudContainer
	if container != nil {
		c = &GCloudContainer{Container: container, Settings: settings}
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

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the GCloud container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
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
func applyOptions(req *testcontainers.GenericContainerRequest, opts []testcontainers.ContainerCustomizer) (options, error) {
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
