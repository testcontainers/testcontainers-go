package blueframevfs

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BlueframeVFSContainer represents the BlueframeVFS container type used in the module
type BlueframeVFSContainer struct {
	testcontainers.Container
	listeningPort string
}

// Run creates an instance of the BlueframeVFS container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*BlueframeVFSContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"80/tcp"},
		Env:          map[string]string{},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("80/tcp"),
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

	// databaseName := req.Env["DATABASE_NAME"]
	// databaseHost := req.Env["DATABASE_HOST"]
	// databasePort := req.Env["DATABASE_PORT"]
	// if databaseName == "" || databaseHost == "" || databasePort == "" {
	// 	return nil, fmt.Errorf("database name, host and port must be provided")
	// }

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *BlueframeVFSContainer
	if container != nil {
		c = &BlueframeVFSContainer{Container: container /*mongoDbHost: databaseHost, mongoDbPort: databasePort, mongoDatabase: databaseName,*/, listeningPort: "9004"}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
