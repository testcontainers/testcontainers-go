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

func (c *Container) ConnectionString(ctx context.Context) (string, error) {
	endpoint, err := c.PortEndpoint(ctx, defaultPort, defaultProtocol)
	if err != nil {
		return "", fmt.Errorf("port endpoint: %w", err)
	}

	// Well-known key for Azure Cosmos DB Emulator
	// See: https://learn.microsoft.com/azure/cosmos-db/emulator
	const testAccKey = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
	return fmt.Sprintf("AccountEndpoint=%s;AccountKey=%s;", endpoint, testAccKey), nil
}
