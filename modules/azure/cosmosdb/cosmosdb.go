package cosmosdb

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultPort     = "8081/tcp"
	defaultProtocol = "http"

	// Well-known, publicly documented account key for the Azure CosmosDB Emulator.
	// See: https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-develop-emulator
	testAccKey = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
)

// Container represents the CosmosDB container type used in the module
type Container struct {
	testcontainers.Container
}

// Run creates an instance of the CosmosDB container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// Initialize with module defaults
	moduleOpts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts(defaultPort),
		testcontainers.WithCmdArgs("--enable-explorer", "false"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForLog("Started"),
				wait.ForListeningPort(nat.Port(defaultPort)),
			),
		),
	}

	// Add user-provided options
	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr}
	}

	if err != nil {
		return c, fmt.Errorf("run cosmosdb: %w", err)
	}

	return c, nil
}

// ConnectionString returns a connection string that can be used to connect to the CosmosDB emulator.
// The connection string includes the account endpoint (host:port) and the default test account key.
// It returns an error if the port endpoint cannot be determined.
//
// Format: "AccountEndpoint=<host>:<port>;AccountKey=<accountKey>"
func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, defaultProtocol)
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	return fmt.Sprintf("AccountEndpoint=%s;AccountKey=%s;", endpoint, testAccKey), nil
}
