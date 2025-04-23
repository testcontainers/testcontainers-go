package dockermodelrunner

import (
	"context"
	"fmt"
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/client"
	"github.com/testcontainers/testcontainers-go/modules/dockermodelrunner/internal/sdk/types"
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

	if settings.model != "" {
		err := c.PullModel(ctx, settings.model)
		if err != nil {
			return c, fmt.Errorf("pull model: %w", err)
		}
	}

	return c, nil
}

// InspectModel returns a model that is already pulled using the Docker Model Runner format.
// The name of the model is in the format of <name>:<tag>.
// The namespace and name defines Models as OCI Artifacts in Docker Hub, therefore the namespace is the organization and the name is the repository.
// E.g. "ai/smollm2:360M-Q4_K_M". See [Models_as_OCI_Artifacts] for more information.
//
// [Models_as_OCI_Artifacts]: https://hub.docker.com/u/ai
func (c *Container) InspectModel(ctx context.Context, namespace string, name string) (*types.ModelResponse, error) {
	dmrClient := client.NewClient(c.baseURL)

	modelResponse, err := dmrClient.InspectModel(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("inspect model: %w", err)
	}

	return modelResponse, nil
}

// ListModels lists all models that are already pulled using the Docker Model Runner format.
func (c *Container) ListModels(ctx context.Context) ([]types.ModelResponse, error) {
	dmrClient := client.NewClient(c.baseURL)

	models, err := dmrClient.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}

	return models, nil
}

// PullModel pulls a model from the Docker Model Runner
func (c *Container) PullModel(ctx context.Context, model string) error {
	log.Default().Printf("🙏 Pulling model %s. Please be patient, no progress bar yet!", model)

	// create a new client for the Docker Model Runner using the OpenAI endpoint
	dmrClient := client.NewClient(c.baseURL)

	_, err := dmrClient.CreateModel(ctx, model)
	if err != nil {
		return fmt.Errorf("create model: %w", err)
	}

	log.Default().Printf("✅ Model %s pulled successfully!", model)

	return nil
}

// OpenAIEndpoint returns the OpenAI endpoint for the Docker Model Runner
func (c *Container) OpenAIEndpoint() string {
	dmrClient := client.NewClient(c.baseURL)

	return dmrClient.OpenAIEndpoint()
}
