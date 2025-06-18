package firestore

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// DefaultProjectID is the default project ID for the Firestore container.
	DefaultProjectID = "test-project"
)

// Container represents the Firestore container type used in the module
type Container struct {
	testcontainers.Container
	settings options
}

// ProjectID returns the project ID of the Firestore container.
func (c *Container) ProjectID() string {
	return c.settings.ProjectID
}

// URI returns the URI of the Firestore container.
func (c *Container) URI() string {
	return c.settings.URI
}

// Run creates an instance of the Firestore GCloud container type.
// The URI uses the empty string as the protocol.
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForListeningPort("8080/tcp"),
			wait.ForLog("running"),
		)),
	}

	settings := defaultOptions()
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, fmt.Errorf("firestore option: %w", err)
			}
		}
	}

	gcloudParameters := "--project=" + settings.ProjectID
	if settings.datastoreMode {
		gcloudParameters += " --database-mode=datastore-mode"
	}

	moduleOpts = append(moduleOpts, testcontainers.WithCmd(
		"/bin/sh",
		"-c",
		"gcloud beta emulators firestore start --host-port 0.0.0.0:8080 "+gcloudParameters,
	))

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if container != nil {
		c = &Container{Container: container, settings: settings}
	}
	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	portEndpoint, err := c.PortEndpoint(ctx, "8080/tcp", "")
	if err != nil {
		return c, fmt.Errorf("port endpoint: %w", err)
	}

	c.settings.URI = portEndpoint

	return c, nil
}
