package azurite

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// EnabledServices is a list of services that should be enabled
	EnabledServices []service
}

func defaultOptions() options {
	return options{
		EnabledServices: []service{blobService, queueService, tableService},
	}
}

// WithInMemoryPersistence is a custom option to enable in-memory persistence for Azurite.
// This option is only available for Azurite v3.28.0 and later.
func WithInMemoryPersistence(megabytes float64) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		cmd := []string{"--inMemoryPersistence"}

		if megabytes > 0 {
			cmd = append(cmd, "--extentMemoryLimit", fmt.Sprintf("%f", megabytes))
		}

		req.Cmd = append(req.Cmd, cmd...)

		return nil
	}
}
