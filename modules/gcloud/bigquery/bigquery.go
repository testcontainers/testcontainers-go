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

	defaultPortNumber9050 = "9050"
	defaultPortNumber9060 = "9060"
	defaultPort9050       = defaultPortNumber9050 + "/tcp"
	defaultPort9060       = defaultPortNumber9060 + "/tcp"
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
		testcontainers.WithExposedPorts(defaultPort9050, defaultPort9060),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort(defaultPort9050),
			wait.ForHTTP("/discovery/v1/apis/bigquery/v2/rest").WithPort(defaultPort9050).WithStatusCodeMatcher(func(status int) bool {
				return status == 200
			}).WithStartupTimeout(time.Second*5),
		),
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("apply option: %w", err)
			}
		}
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs("--project", settings.ProjectID))

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("run bigquery: %w", err)
	}

	portEndpoint, err := c.PortEndpoint(ctx, defaultPort9050, "http")
	if err != nil {
		return c, fmt.Errorf("port endpoint: %w", err)
	}

	c.settings.URI = portEndpoint

	return c, nil
}
