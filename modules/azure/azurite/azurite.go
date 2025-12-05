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

// ServiceURL returns the URL of the given service
func (c *Container) ServiceURL(ctx context.Context, srv Service) (string, error) {
	port, err := servicePort(srv)
	if err != nil {
		return "", err
	}

	return c.PortEndpoint(ctx, port, "http")
}

// BlobServiceURL returns the URL of the Blob service
func (c *Container) BlobServiceURL(ctx context.Context) (string, error) {
	return c.ServiceURL(ctx, BlobService)
}

// QueueServiceURL returns the URL of the Queue service
func (c *Container) QueueServiceURL(ctx context.Context) (string, error) {
	return c.ServiceURL(ctx, QueueService)
}

// TableServiceURL returns the URL of the Table service
func (c *Container) TableServiceURL(ctx context.Context) (string, error) {
	return c.ServiceURL(ctx, TableService)
}

func servicePort(srv Service) (nat.Port, error) {
	switch srv {
	case BlobService:
		return BlobPort, nil
	case QueueService:
		return QueuePort, nil
	case TableService:
		return TablePort, nil
	default:
		return "", fmt.Errorf("unknown service: %s", srv)
	}
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
	if len(settings.EnabledServices) == 1 && settings.EnabledServices[0] != TableService {
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
			port, err := servicePort(srv)
			if err != nil {
				return nil, err
			}

			cmd = append(cmd, fmt.Sprintf("--%sHost", srv), "0.0.0.0", fmt.Sprintf("--%sPort", srv), port.Port())
			exposedPorts = append(exposedPorts, string(port))
			waitingFor = append(waitingFor, wait.ForListeningPort(port))
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
