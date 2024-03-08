package weaviate

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// WeaviateContainer represents the Weaviate container type used in the module
type WeaviateContainer struct {
	testcontainers.Container
}

// RunContainer creates an instance of the Weaviate container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*WeaviateContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "semitechnologies/weaviate:1.24.1",
		Cmd:          []string{"--host", "0.0.0.0", "--scheme", "http", "--port", "8080"},
		ExposedPorts: []string{"8080/tcp", "50051/tcp"},
		Env: map[string]string{
			"QUERY_DEFAULTS_LIMIT":                    "25",
			"AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED": "true",
			"PERSISTENCE_DATA_PATH":                   "/var/lib/weaviate",
			"DEFAULT_VECTORIZER_MODULE":               "none",
			"ENABLE_MODULES":                          "text2vec-cohere,text2vec-huggingface,text2vec-palm,text2vec-openai,generative-openai,generative-cohere,generative-palm,ref2vec-centroid,reranker-cohere,qna-openai",
			"CLUSTER_HOSTNAME":                        "node1",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8080").WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("50051").WithStartupTimeout(5*time.Second),
		),
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

	return &WeaviateContainer{Container: container}, nil
}

// HttpHostAddress returns the schema and host of the Weaviate container.
// At the moment, it only supports the http scheme.
func (c *WeaviateContainer) HttpHostAddress(ctx context.Context) (string, string, error) {
	containerPort, err := c.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return "", "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container host")
	}

	return "http", fmt.Sprintf("%s:%s", host, containerPort.Port()), nil
}

// GrpcHostAddress returns the gRPC host of the Weaviate container.
// At the moment, it only supports unsecured gRPC connection.
func (c *WeaviateContainer) GrpcHostAddress(ctx context.Context) (string, error) {
	containerPort, err := c.MappedPort(ctx, "50051/tcp")
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get container host")
	}

	return fmt.Sprintf("%s:%s", host, containerPort.Port()), nil
}
