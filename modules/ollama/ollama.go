package ollama

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
)

// OllamaContainer represents the Ollama container type used in the module
type OllamaContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Ollama container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*OllamaContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "ollama/ollama:0.1.25",
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		opt.Customize(&genericContainerReq)
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &OllamaContainer{Container: container}, nil
}
