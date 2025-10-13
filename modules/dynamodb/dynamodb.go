package dynamodb

import (
	"context"
	"fmt"

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

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *DynamoDBContainer
	if ctr != nil {
		c = &DynamoDBContainer{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run dynamodb: %w", err)
	}

	return c, nil
}

// ConnectionString returns DynamoDB local endpoint host and port in <host>:<port> format
func (c *DynamoDBContainer) ConnectionString(ctx context.Context) (string, error) {
	return c.PortEndpoint(ctx, port, "")
}

// WithSharedDB allows container reuse between successive runs. Data will be persisted
func WithSharedDB() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		err := testcontainers.WithCmdArgs("-sharedDb")(req)
		if err != nil {
			return fmt.Errorf("with shared db: %w", err)
		}

		err = testcontainers.WithReuseByName(containerName)(req)
		if err != nil {
			return fmt.Errorf("with reuse by name: %w", err)
		}

		return nil
	}
}

// WithDisableTelemetry - DynamoDB local will not send any telemetry
func WithDisableTelemetry() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// if other flags (e.g. -sharedDb) exist, append to them
		return testcontainers.WithCmdArgs("-disableTelemetry")(req)
	}
}
