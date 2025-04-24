package dockermodelrunner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/client"
	"github.com/testcontainers/testcontainers-go/modules/socat"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	modelRunnerEntrypoint = "model-runner.docker.internal"
	modelRunnerPort       = 80
)

// Container represents the DockerModelRunner container type used in the module
type Container struct {
	*socat.Container
	*client.Client
	model   string
	baseURL string
}

// Run creates an instance of the DockerModelRunner container type.
func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	settings := defaultOptions()

	// Process model runner options.
	for _, opt := range opts {
		if apply, ok := opt.(Option); ok {
			if err := apply(&settings); err != nil {
				return nil, err
			}
		}
	}

	// Add socat options, which are applied to the socat container.
	opts = append(opts, testcontainers.WithWaitStrategy(
		wait.ForListeningPort("80/tcp"),
		wait.ForHTTP("/").WithPort("80/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}),
	))
	opts = append(opts, socat.WithTarget(socat.NewTarget(modelRunnerPort, modelRunnerEntrypoint)))

	socatCtr, err := socat.Run(ctx, socat.DefaultImage, opts...)
	var c *Container
	if socatCtr != nil {
		c = &Container{Container: socatCtr, model: settings.model}
	}

	if err != nil {
		return c, fmt.Errorf("socat run: %w", err)
	}

	c.baseURL = socatCtr.TargetURL(modelRunnerPort).String()

	c.Client = client.NewClient(c.baseURL)

	if settings.model != "" {
		err := c.PullModel(ctx, settings.model)
		if err != nil {
			return c, fmt.Errorf("pull model: %w", err)
		}
	}

	return c, nil
}
