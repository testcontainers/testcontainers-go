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
	// fabricPort is the port used for Intra-cluster communication port.
	// Replica writes, migrations, and other node-to-node communications use the Fabric port.
	fabricPort = "3001/tcp"
	// heartbeatPort is the port used for heartbeat communication
	// between nodes in the Aerospike cluster
	heartbeatPort = "3002/tcp"
	// infoPort is the port used for info commands
	infoPort = "3003/tcp"
)

// Container is the Aerospike container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the Aerospike container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{port, fabricPort, heartbeatPort, infoPort},
		Env: map[string]string{
			"AEROSPIKE_CONFIG_FILE": "/etc/aerospike/aerospike.conf",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("migrations: complete"),
			wait.ForListeningPort(port).WithStartupTimeout(10*time.Second),
			wait.ForListeningPort(fabricPort).WithStartupTimeout(10*time.Second),
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
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
