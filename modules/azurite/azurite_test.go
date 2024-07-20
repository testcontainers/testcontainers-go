package azurite_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
)

func TestAzurite(t *testing.T) {
	ctx := context.Background()

	container, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	// perform assertions
}
