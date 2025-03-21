// Deprecated: This package is deprecated in favor of "modules/azure/azurite".
// Please use that package instead for all new code.
package azurite

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
)

const (
	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	BlobPort = azurite.BlobPort
	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	QueuePort = azurite.QueuePort
	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	TablePort = azurite.TablePort

	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	// AccountName is the default testing account name used by Azurite
	AccountName string = azurite.AccountName

	// Deprecated: This constant is deprecated in favor of the one in "modules/azure/azurite".
	// Please use that package instead for all new code.
	// AccountKey is the default testing account key used by Azurite
	AccountKey string = azurite.AccountKey
)

// Deprecated: This type is deprecated in favor of the one in "modules/azure/azurite".
// AzuriteContainer represents the Azurite container type used in the module
type AzuriteContainer = azurite.Container

// Deprecated: This function is deprecated in favor of the one in "modules/azure/azurite".
// RunContainer creates an instance of the Azurite container type
func RunContainer(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*AzuriteContainer, error) {
	return Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.28.0", opts...)
}

// Deprecated: This function is deprecated in favor of the one in "modules/azure/azurite".
// Run creates an instance of the Azurite container type
func Run(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AzuriteContainer, error) {
	return azurite.Run(ctx, img, opts...)
}
