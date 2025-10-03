package bigquery

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultProjectID is the default project ID for the BigQuery container.
	DefaultProjectID = "test-project"

	// bigQueryDataYamlPath is the path to the data yaml file in the container.
	bigQueryDataYamlPath = "/testcontainers-data.yaml"
)

// Container represents the BigQuery container type used in the module
type Container struct {
	testcontainers.Container
	settings options
}

// ProjectID returns the project ID of the BigQuery container.
func (c *Container) ProjectID() string {
	return c.settings.ProjectID
}

// URI returns the URI of the BigQuery container.
func (c *Container) URI() string {
	return c.settings.URI
}

// Run creates an instance of the BigQuery GCloud container type.
// The URI uses http:// as the protocol.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("9050/tcp", "9060/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("9050/tcp"),
			wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort("9050/tcp").WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}).WithStartupTimeout(time.Second*5),
		)),
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("apply option: %w", err)
			}
		}
	}

	moduleOpts = append(moduleOpts, opts...)

	moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs("--project", settings.ProjectID))

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("run bigquery: %w", err)
	}

	portEndpoint, err := c.PortEndpoint(ctx, "9050/tcp", "http")
	if err != nil {
		return c, fmt.Errorf("port endpoint: %w", err)
	}

	c.settings.URI = portEndpoint

	return c, nil
}
