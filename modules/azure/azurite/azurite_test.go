package azurite_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/azurite"
)

func TestAzurite(t *testing.T) {
	ctx := context.Background()

	ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

func TestAzurite_inMemoryPersistence(t *testing.T) {
	ctx := context.Background()

	t.Run("v28-above", func(t *testing.T) {
		ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.28.0", azurite.WithInMemoryPersistence(64))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)
	})

	t.Run("v27-below", func(t *testing.T) {
		ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0", azurite.WithInMemoryPersistence(64))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
	})
}

func TestAzurite_serviceURL(t *testing.T) {
	ctx := context.Background()

	ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("blob", func(t *testing.T) {
		_, err := ctr.ServiceURL(ctx, azurite.BlobService)
		require.NoError(t, err)
	})

	t.Run("queue", func(t *testing.T) {
		_, err := ctr.ServiceURL(ctx, azurite.QueueService)
		require.NoError(t, err)
	})

	t.Run("table", func(t *testing.T) {
		_, err := ctr.ServiceURL(ctx, azurite.TableService)
		require.NoError(t, err)
	})

	t.Run("unknown-service", func(t *testing.T) {
		_, err := ctr.ServiceURL(ctx, "unknown")
		require.Error(t, err)
	})
}
