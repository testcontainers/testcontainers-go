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
func newGCloudContainer(ctx context.Context, port int, c testcontainers.Container, settings options) (*GCloudContainer, error) {
	mappedPort, err := c.MappedPort(ctx, nat.Port(fmt.Sprintf("%d/tcp", port)))
	if err != nil {
		return nil, err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	gCloudContainer := &GCloudContainer{
		Container: c,
		Settings:  settings,
		URI:       uri,
	}

	return gCloudContainer, nil
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
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithProjectID sets the project ID for the GCloud container.
func WithProjectID(projectID string) Option {
	return func(o *options) {
		o.ProjectID = projectID
	}
}

// applyOptions applies the options to the container request and returns the settings.
func applyOptions(req testcontainers.GenericContainerRequest, opts []testcontainers.ContainerCustomizer) options {
	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			apply(&settings)
		}
		opt.Customize(&req)
	}

	return settings
}
