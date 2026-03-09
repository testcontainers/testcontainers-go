package shared

import (
	"github.com/testcontainers/testcontainers-go"
)

const (
	// DefaultProjectID is the default project ID for the Pubsub container.
	DefaultProjectID = "test-project"
)

// Options represents the options for the different GCloud containers.
// This type must contain all the options that are common to all the GCloud containers.
type Options struct {
	ProjectID string
	URI       string
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the GCloud container.
type Option func(*Options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// DefaultOptions returns a new Options instance with the default project ID.
func DefaultOptions() Options {
	return Options{
		ProjectID: DefaultProjectID,
	}
}

// WithProjectID sets the project ID for the GCloud container.
func WithProjectID(projectID string) Option {
	return func(o *Options) error {
		o.ProjectID = projectID
		return nil
	}
}
