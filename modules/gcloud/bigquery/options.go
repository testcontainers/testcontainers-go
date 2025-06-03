package bigquery

import (
	"errors"
	"io"
	"slices"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/gcloud/internal/shared"
)

// Options aliases the common GCloud options
type options = shared.Options

// Option aliases the common GCloud option type
type Option = shared.Option

// defaultOptions returns a new Options instance with the default project ID.
func defaultOptions() options {
	return shared.DefaultOptions()
}

// WithProjectID re-exports the common GCloud WithProjectID option
var WithProjectID = shared.WithProjectID

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
