package gcloud

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

const defaultProjectID = "test-project"

type GCloudContainer interface {
	uri(ctx context.Context) (string, error)
}

func containerURI(ctx context.Context, container GCloudContainer) (string, error) {
	return container.uri(ctx)
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

// Option is an option for the Redpanda container.
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
