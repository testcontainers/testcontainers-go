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
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{string(port)},
		Entrypoint:   []string{"java", "-Djava.library.path=./DynamoDBLocal_lib"},
		Cmd:          []string{"-jar", "DynamoDBLocal.jar"},
		WaitingFor:   wait.ForListeningPort(port),
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *DynamoDBContainer
	if container != nil {
		c = &DynamoDBContainer{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
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

// WithSharedDB allows container reuse between successive runs. Data will be persisted
func WithSharedDB() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "-sharedDb")

		req.Reuse = true
		req.Name = containerName

		return nil
	}
}

// WithDisableTelemetry - DynamoDB local will not send any telemetry
func WithDisableTelemetry() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		// if other flags (e.g. -sharedDb) exist, append to them
		req.Cmd = append(req.Cmd, "-disableTelemetry")

		return nil
	}
}
