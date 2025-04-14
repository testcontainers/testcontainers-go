package dockermodelrunner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/sdk/client"
	"github.com/testcontainers/testcontainers-go/modules/socat"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	modelRunnerEntrypoint = "model-runner.docker.internal"
	modelRunnerPort       = 80
	openAIEndpointSuffix  = "/engines/v1"
)

// Container represents the DockerModelRunner container type used in the module
type Container struct {
	*socat.Container
	model          string
	openAIEndpoint string
}

// Run creates an instance of the DockerModelRunner container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
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
		}).WithResponseMatcher(func(body io.Reader) bool {
			bodyBytes, err := io.ReadAll(body)
			if err != nil {
				return false
			}
			return strings.Contains(string(bodyBytes), "The service is running")
		}),
	))
	opts = append(opts, socat.WithTarget(socat.NewTarget(modelRunnerPort, modelRunnerEntrypoint)))

	socatCtr, err := socat.Run(ctx, img, opts...)
	var c *Container
	if socatCtr != nil {
		c = &Container{Container: socatCtr, model: settings.model}
	}

	if err != nil {
		return c, fmt.Errorf("socat run: %w", err)
	}

	c.openAIEndpoint = socatCtr.TargetURL(modelRunnerPort).String() + openAIEndpointSuffix

	if settings.model != "" {
		err := c.PullModel(ctx, settings.model)
		if err != nil {
			return c, fmt.Errorf("pull model: %w", err)
		}
	}

	return c, nil
}

// PullModel pulls a model from the Docker Model Runner
func (c *Container) PullModel(ctx context.Context, model string) error {
	log.Default().Printf("🙏 Pulling model %s. Please be patient, no progress bar yet!", model)

	// create a new client for the Docker Model Runner using the OpenAI endpoint
	dmrClient := client.NewClient(c.openAIEndpoint)

	_, err := dmrClient.CreateModel(ctx, model)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", model)

	return nil
}

// OpenAIEndpoint returns the OpenAI endpoint for the Docker Model Runner
func (c *Container) OpenAIEndpoint(ctx context.Context) string {
	return c.openAIEndpoint
}
