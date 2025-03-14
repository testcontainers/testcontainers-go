package azurite

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
)

// Deprecated: This option is deprecated in favor of the one in "modules/azure/azurite".
// Please use that package instead for all new code.
// WithInMemoryPersistence is a custom option to enable in-memory persistence for Azurite.
// This option is only available for Azurite v3.28.0 and later.
func WithInMemoryPersistence(megabytes float64) testcontainers.CustomizeRequestOption {
	return azurite.WithInMemoryPersistence(megabytes)
}
