package bigquery

import (
	"errors"
	"io"
	"slices"

	"github.com/testcontainers/testcontainers-go"
)

// Options represents the options for the different GCloud containers.
// This type must contain all the options that are common to all the GCloud containers.
type options struct {
	ProjectID string
	URI       string
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the GCloud container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// defaultOptions returns a new Options instance with the default project ID.
func defaultOptions() options {
	return options{
		ProjectID: DefaultProjectID,
	}
}

// WithProjectID sets the project ID for the GCloud container.
func WithProjectID(projectID string) Option {
	return func(o *options) error {
		o.ProjectID = projectID
		return nil
	}
}

// WithDataYAML seeds the BigQuery project for the GCloud container with an [io.Reader] representing
// the data yaml file, which is used to copy the file to the container, and then processed to seed
// the BigQuery project.
//
// Other GCloud containers will ignore this option.
func WithDataYAML(r io.Reader) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		if slices.Contains(req.Cmd, "--data-from-yaml") {
			return errors.New("data yaml already exists")
		}

		req.Cmd = append(req.Cmd, "--data-from-yaml", bigQueryDataYamlPath)

		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            r,
			ContainerFilePath: bigQueryDataYamlPath,
			FileMode:          0o644,
		})

		return nil
	}
}
