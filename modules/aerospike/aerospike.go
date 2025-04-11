package aerospike

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// port is the port used for client connections
	port = "3000/tcp"
	// febricPort is the port used for fabric communication
	febricPort = "3001/tcp"
	// heartbeatPort is the port used for heartbeat communication
	// between nodes in the Aerospike cluster
	heartbeatPort = "3002/tcp"
	// infoPort is the port used for info commands
	infoPort = "3003/tcp"
)

// AerospikeContainer is the Aerospike container type used in the module
type AerospikeContainer struct {
	testcontainers.Container
}

// Deprecated: use Run instead
// RunContainer creates an instance of the Aerospike container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*AerospikeContainer, error) {
	return Run(ctx, "aerospike/aerospike-server:latest", opts...)
}

// Run creates an instance of the Aerospike container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AerospikeContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{port, febricPort, heartbeatPort, infoPort},
		Env: map[string]string{
			"AEROSPIKE_CONFIG_FILE": "/etc/aerospike/aerospike.conf",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("migrations: complete"),
			wait.ForListeningPort(port).WithStartupTimeout(10*time.Second),
			wait.ForListeningPort(febricPort).WithStartupTimeout(10*time.Second),
			wait.ForListeningPort(heartbeatPort).WithStartupTimeout(10*time.Second),
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

	return &AerospikeContainer{
		Container: container,
	}, nil
}

// GetHost returns the host of the Aerospike container
func (c *AerospikeContainer) GetHost() (string, error) {

	host, err := c.Host(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get container host: %w", err)
	}
	return host, nil
}

// GetPort returns the port of the Aerospike container
func (c *AerospikeContainer) GetPort() (string, error) {
	port, err := c.MappedPort(context.Background(), port)
	if err != nil {
		return "", fmt.Errorf("failed to get container port: %w", err)
	}
	return port.Port(), nil
}
