package ollama

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deprecated: it will be removed in the next major version.
const DefaultOllamaImage = "ollama/ollama:0.5.7"

// OllamaContainer represents the Ollama container type used in the module
type OllamaContainer struct {
	testcontainers.Container
}

// ConnectionString returns the connection string for the Ollama container,
// using the default port 11434.
func (c *OllamaContainer) ConnectionString(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	port, err := c.MappedPort(ctx, "11434/tcp")
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%d", host, port.Int()), nil
}

// Commit it commits the current file system changes in the container into a new target image.
// The target image name should be unique, as this method will commit the current state
// of the container into a new image with the given name, so it doesn't override existing images.
// It should be used for creating an image that contains a loaded model.
func (c *OllamaContainer) Commit(ctx context.Context, targetImage string) error {
	if _, ok := c.Container.(*localProcess); ok {
		return nil
	}

	cli, err := testcontainers.NewDockerClientWithOpts(context.Background())
	if err != nil {
		return err
	}

	list, err := cli.ImageList(ctx, image.ListOptions{Filters: filters.NewArgs(filters.Arg("reference", targetImage))})
	if err != nil {
		return fmt.Errorf("listing images %w", err)
	}

	if len(list) > 0 {
		return fmt.Errorf("image %s already exists", targetImage)
	}

	_, err = cli.ContainerCommit(ctx, c.GetContainerID(), container.CommitOptions{
		Reference: targetImage,
		Config: &container.Config{
			Labels: map[string]string{
				core.LabelSessionID: "",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("committing container %w", err)
	}

	return nil
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Ollama container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OllamaContainer, error) {
	return Run(ctx, "ollama/ollama:0.5.7", opts...)
}

// Run creates an instance of the Ollama container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*OllamaContainer, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        img,
			ExposedPorts: []string{"11434/tcp"},
			WaitingFor:   wait.ForListeningPort("11434/tcp").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	}

	// Always request a GPU if the host supports it.
	opts = append(opts, withGpu())

	var local *localProcess
	for _, opt := range opts {
		if err := opt.Customize(&req); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
		if l, ok := opt.(*localProcess); ok {
			local = l
		}
	}

	// Now we have processed all the options, we can check if we need to use the local process.
	if local != nil {
		return local.run(ctx, req)
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	var c *OllamaContainer
	if container != nil {
		c = &OllamaContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
