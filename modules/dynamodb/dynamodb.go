package dynamodb

import (
	"context"
	"fmt"
	"slices"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port          = "8000/tcp"
	containerName = "tc_dynamodb_local"
)

// DynamoDBContainer represents the DynamoDB container type used in the module
type DynamoDBContainer struct {
	testcontainers.Container
}

// Run creates an instance of the DynamoDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*DynamoDBContainer, error) {
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithEntrypoint("java", "-Djava.library.path=./DynamoDBLocal_lib"),
		testcontainers.WithCmd("-jar", "DynamoDBLocal.jar"),
		testcontainers.WithExposedPorts(port),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(port)),
	}

	moduleOpts = append(moduleOpts, opts...)

	defaultOptions := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&defaultOptions); err != nil {
				return nil, fmt.Errorf("dynamodb option: %w", err)
			}
		}
	}

	if slices.Contains(defaultOptions.cmd, "-sharedDb") {
		moduleOpts = append(moduleOpts, testcontainers.WithReuseByName(containerName))
	}

	// module options take precedence over default options
	moduleOpts = append(moduleOpts, testcontainers.WithCmdArgs(defaultOptions.cmd...))

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *DynamoDBContainer
	if container != nil {
		c = &DynamoDBContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("run: %w", err)
	}

	return c, nil
}

// ConnectionString returns DynamoDB local endpoint host and port in <host>:<port> format
func (c *DynamoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	hostIP, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	return hostIP + ":" + mappedPort.Port(), nil
}
