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
	return c.PortEndpoint(ctx, "11434/tcp", "http")
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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("11434/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("11434/tcp").WithStartupTimeout(60 * time.Second)),
	}

	moduleOpts = append(moduleOpts, opts...)

	// Always request a GPU if the host supports it.
	moduleOpts = append(moduleOpts, withGpu())

	var local *localProcess
	for _, opt := range opts {
		if l, ok := opt.(*localProcess); ok {
			local = l
		}
	}

	// Now we have processed all the options, we can check if we need to use the local process.
	if local != nil {
		// pass the image to the local process
		moduleOpts = append(moduleOpts, testcontainers.WithImage(img))
		return local.run(ctx, moduleOpts...)
	}

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *OllamaContainer
	if ctr != nil {
		c = &OllamaContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}
