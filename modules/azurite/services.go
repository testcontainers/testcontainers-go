package azurite

import (
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
)

const (
	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	//
	//nolint:staticcheck
	BlobService Service = azurite.BlobService

	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	//
	//nolint:staticcheck
	QueueService Service = azurite.QueueService

	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	//
	//nolint:staticcheck
	TableService Service = azurite.TableService
)

// Deprecated: This type is deprecated in favor of the one in "modules/azure/azurite".
// Please use that package instead for all new code.
//
//nolint:staticcheck
type Service = azurite.Service
