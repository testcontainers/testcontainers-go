package azurite

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// EnabledServices is a list of services that should be enabled
	EnabledServices []Service
}

func defaultOptions() options {
	return options{
		EnabledServices: []Service{BlobService, QueueService, TableService},
	}
}

// Satisfy the testcontainers.ContainerCustomizer interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Azurite container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithEnabledServices is a custom option to specify which services should be enabled.
func WithEnabledServices(services ...Service) Option {
	return func(o *options) error {
		if len(services) == 0 {
			services = []Service{BlobService, QueueService, TableService}
		} else {
			seen := make(map[Service]bool, len(services))
			for _, srv := range services {
				if seen[srv] {
					return fmt.Errorf("duplicate service: %s", srv)
				}
				seen[srv] = true

				switch srv {
				case BlobService, QueueService, TableService:
					// valid service, continue
				default:
					return fmt.Errorf("unknown service: %s", srv)
				}
			}
		}

		o.EnabledServices = services
		return nil
	}
}

// WithInMemoryPersistence is a custom option to enable in-memory persistence for Azurite.
// This option is only available for Azurite v3.28.0 and later.
func WithInMemoryPersistence(megabytes float64) testcontainers.CustomizeRequestOption {
	cmd := []string{"--inMemoryPersistence"}

	if megabytes > 0 {
		cmd = append(cmd, "--extentMemoryLimit", fmt.Sprintf("%f", megabytes))
	}

	return testcontainers.WithCmdArgs(cmd...)
}
