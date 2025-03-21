package azurite

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// BlobPort is the default port used by Azurite
	BlobPort = "10000/tcp"
	// QueuePort is the default port used by Azurite
	QueuePort = "10001/tcp"
	// TablePort is the default port used by Azurite
	TablePort = "10002/tcp"

	// defaultCredentials {
	// AccountName is the default testing account name used by Azurite
	AccountName string = "devstoreaccount1"

	// AccountKey is the default testing account key used by Azurite
	AccountKey string = "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
	// }
)

// Container represents the Azurite container type used in the module
type Container struct {
	testcontainers.Container
	opts options
}

// Deprecated: Use [azure.BlobServiceURL], [azure.QueueServiceURL], or [azure.TableServiceURL] methods instead.
// ServiceURL returns the URL of the given service
func (c *Container) ServiceURL(ctx context.Context, srv Service) (string, error) {
	return c.serviceURL(ctx, srv)
}

// BlobServiceURL returns the URL of the Blob service
func (c *Container) BlobServiceURL(ctx context.Context) (string, error) {
	return c.serviceURL(ctx, blobService)
}

// QueueServiceURL returns the URL of the Queue service
func (c *Container) QueueServiceURL(ctx context.Context) (string, error) {
	return c.serviceURL(ctx, queueService)
}

// TableServiceURL returns the URL of the Table service
func (c *Container) TableServiceURL(ctx context.Context) (string, error) {
	return c.serviceURL(ctx, tableService)
}

func (c *Container) serviceURL(ctx context.Context, srv service) (string, error) {
	hostname, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("host: %w", err)
	}

	var port nat.Port
	switch srv {
	case blobService:
		port = BlobPort
	case queueService:
		port = QueuePort
	case tableService:
		port = TablePort
	default:
		return "", fmt.Errorf("unknown service: %s", srv)
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", fmt.Errorf("mapped port: %w", err)
	}

	return fmt.Sprintf("http://%s:%d", hostname, mappedPort.Int()), nil
}

// Run creates an instance of the Azurite container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{BlobPort, QueuePort, TablePort},
		Env:          map[string]string{},
		Entrypoint:   []string{"azurite"},
		Cmd:          []string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// 1. Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, fmt.Errorf("customize: %w", err)
		}
	}

	// 2. evaluate the enabled services to apply the right wait strategy and Cmd options
	if len(settings.EnabledServices) > 0 {
		waitingFor := make([]wait.Strategy, 0, len(settings.EnabledServices))
		for _, srv := range settings.EnabledServices {
			switch srv {
			case BlobService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--blobHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForListeningPort(BlobPort))
			case QueueService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--queueHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForListeningPort(QueuePort))
			case TableService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--tableHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForListeningPort(TablePort))
			}
		}

		if genericContainerReq.WaitingFor != nil {
			genericContainerReq.WaitingFor = wait.ForAll(genericContainerReq.WaitingFor, wait.ForAll(waitingFor...))
		} else {
			genericContainerReq.WaitingFor = wait.ForAll(waitingFor...)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	var c *Container
	if container != nil {
		c = &Container{Container: container, opts: settings}
	}

	if err != nil {
		return c, fmt.Errorf("generic container: %w", err)
	}

	return c, nil
}
