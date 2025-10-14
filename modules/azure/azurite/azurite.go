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

	return c.PortEndpoint(ctx, port, "http")
}

// Run creates an instance of the Azurite container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*Container, error) {
	// 1. Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if o, ok := opt.(Option); ok {
			if err := o(&settings); err != nil {
				return nil, fmt.Errorf("azurite option: %w", err)
			}
		}
	}

	entrypoint := "azurite"
	if len(settings.EnabledServices) == 1 && settings.EnabledServices[0] != tableService {
		// Use azurite-table in future once it matures. Graceful shutdown is currently very slow.
		entrypoint = fmt.Sprintf("%s-%s", entrypoint, settings.EnabledServices[0])
	}
	moduleOpts := []testcontainers.ContainerCustomizer{testcontainers.WithEntrypoint(entrypoint)}

	// 2. evaluate the enabled services to apply the right wait strategy and Cmd options
	if len(settings.EnabledServices) > 0 {
		cmd := make([]string, 0, len(settings.EnabledServices))
		exposedPorts := make([]string, 0, len(settings.EnabledServices))
		waitingFor := make([]wait.Strategy, 0, len(settings.EnabledServices))
		for _, srv := range settings.EnabledServices {
			switch srv {
			case BlobService:
				cmd = append(cmd, "--blobHost", "0.0.0.0", "--blobPort", BlobPort)
				exposedPorts = append(exposedPorts, BlobPort)
				waitingFor = append(waitingFor, wait.ForListeningPort(BlobPort))
			case QueueService:
				cmd = append(cmd, "--queueHost", "0.0.0.0", "--queuePort", QueuePort)
				exposedPorts = append(exposedPorts, QueuePort)
				waitingFor = append(waitingFor, wait.ForListeningPort(QueuePort))
			case TableService:
				cmd = append(cmd, "--tableHost", "0.0.0.0", "--tablePort", TablePort)
				exposedPorts = append(exposedPorts, TablePort)
				waitingFor = append(waitingFor, wait.ForListeningPort(TablePort))
			}
		}

		moduleOpts = append(moduleOpts,
			testcontainers.WithCmd(cmd...),
			testcontainers.WithExposedPorts(exposedPorts...),
			testcontainers.WithWaitStrategy(waitingFor...),
		)
	}

	moduleOpts = append(moduleOpts, opts...)

	ctr, err := testcontainers.Run(ctx, img, moduleOpts...)
	var c *Container
	if ctr != nil {
		c = &Container{Container: ctr, opts: settings}
	}

	if err != nil {
		return c, fmt.Errorf("run azurite: %w", err)
	}

	return c, nil
}
