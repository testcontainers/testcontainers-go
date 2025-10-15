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

func TestAzurite_enabledServices(t *testing.T) {
	ctx := context.Background()

	services := []azurite.Service{azurite.BlobService, azurite.QueueService, azurite.TableService, "invalid"}
	for _, service := range services {
		t.Run(string(service), func(t *testing.T) {
			ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.33.0", azurite.WithInMemoryPersistence(0), azurite.WithEnabledServices(service))
			testcontainers.CleanupContainer(t, ctr)
			if service == "invalid" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			for _, srv := range services {
				_, err = ctr.ServiceURL(ctx, srv)
				if srv == service {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			}
		})
	}
}

func TestAzurite_serviceURL(t *testing.T) {
	ctx := context.Background()

	ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("blob", func(t *testing.T) {
		_, err := ctr.BlobServiceURL(ctx)
		require.NoError(t, err)
	})

	t.Run("queue", func(t *testing.T) {
		_, err := ctr.QueueServiceURL(ctx)
		require.NoError(t, err)
	})

	t.Run("table", func(t *testing.T) {
		_, err := ctr.TableServiceURL(ctx)
		require.NoError(t, err)
	})
}
