package aerospike

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	port          = "3000/tcp"
	containerName = "tc_dynamodb_local"
)

type AerospikeContainer struct {
	testcontainers.Container
	Host string
	Port int
}

// ConnectionHost returns the host and port of the cassandra container, using the default, native 9000 port, and
// obtaining the host and exposed port from the container
// func (c *AerospikeContainer) ConnectionHost(ctx context.Context) (string, error) {
// 	host, err := c.Host(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	port, err := c.MappedPort(ctx, port)
// 	if err != nil {
// 		return "", err
// 	}

// 	return host + ":" + port.Port(), nil
// }

// Run creates an instance of the Aerospike container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AerospikeContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"3000/tcp", "3001/tcp", "3002/tcp", "3003/tcp"},
		Env: map[string]string{
			"AEROSPIKE_CONFIG_FILE": "/etc/aerospike/aerospike.conf",
			"NAMESPACE":             "test",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("migrations: complete"),
			wait.ForListeningPort("3000/tcp").WithStartupTimeout(10*time.Second),
			wait.ForListeningPort("3001/tcp").WithStartupTimeout(10*time.Second),
			wait.ForListeningPort("3002/tcp").WithStartupTimeout(10*time.Second),
		),
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
	if err != nil {
		return nil, fmt.Errorf("failed to start Aerospike container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, port)
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	return &AerospikeContainer{
		Container: container,
		Host:      host,
		Port:      port.Int(),
	}, nil
}
