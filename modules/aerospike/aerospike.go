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
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(port, fabricPort, heartbeatPort, infoPort),
		testcontainers.WithEnv(map[string]string{
			"AEROSPIKE_CONFIG_FILE": "/etc/aerospike/aerospike.conf",
		}),
		testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForLog("migrations: complete"),
			wait.ForListeningPort(port).WithStartupTimeout(10*time.Second),
			wait.ForListeningPort(fabricPort).WithStartupTimeout(10*time.Second),
			wait.ForListeningPort(heartbeatPort).WithStartupTimeout(10*time.Second),
		)),
	}

	moduleOpts = append(moduleOpts, opts...)

	container, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if container != nil {
		c = &Container{Container: container}
	}

	if err != nil {
		return c, fmt.Errorf("run aerospike: %w", err)
	}

	return c, nil
}
