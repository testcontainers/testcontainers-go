package bigtable

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultProjectID is the default project ID for the BigTable container.
	DefaultProjectID = "test-project"
)

// Container represents the BigTable container type used in the module
type Container struct {
	testcontainers.Container
	settings options
}

// ProjectID returns the project ID of the BigTable container.
func (c *Container) ProjectID() string {
	return c.settings.ProjectID
}

// URI returns the URI of the BigTable container.
func (c *Container) URI() string {
	return c.settings.URI
}

// Run creates an instance of the BigTable GCloud container type.
// The URI uses the empty string as the protocol.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"9000/tcp"},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("9000/tcp"),
				wait.ForLog("running"),
			),
		},
		Started: true,
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
		if err := opt.Customize(&req); err != nil {
			return nil, err
		}
	}

	req.Cmd = []string{
		"/bin/sh",
		"-c",
		"gcloud beta emulators bigtable start --host-port 0.0.0.0:9000 --project=" + settings.ProjectID,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	var c *Container
	if container != nil {
		c = &Container{Container: container, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	portEndpoint, err := c.PortEndpoint(ctx, "9000/tcp", "")
	if err != nil {
		return c, fmt.Errorf("port endpoint: %w", err)
	}

	c.settings.URI = portEndpoint

	return c, nil
}
