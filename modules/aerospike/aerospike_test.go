package aerospike_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	as "github.com/testcontainers/testcontainers-go/modules/aerospike"
)

// TestAerospikeDB tests the AerospikeDB container functionality
// It includes tests for starting the container with a valid image,
// applying container customizations, and handling context cancellation.
// It also includes a test for an invalid image to ensure proper error handling.
// The tests use the testcontainers-go library to manage container lifecycle
func TestAeroSpikeDB(t *testing.T) {
	t.Run("fails with invalid image", func(t *testing.T) {
		ctx := context.Background()
		_, err := as.Run(ctx, "invalid-aerospike-image")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to start Aerospike container")
	})

	t.Run("succeeds with valid image", func(t *testing.T) {
		ctx := context.Background()
		container, err := as.Run(ctx, "aerospike/aerospike-server:latest")
		require.NoError(t, err)
		require.NotNil(t, container)
		defer container.Container.Terminate(ctx)

		require.NotEmpty(t, container.Host)
		require.NotEmpty(t, container.Port)
	})

	t.Run("applies container customizations", func(t *testing.T) {
		ctx := context.Background()
		customEnv := "TEST_ENV=value"
		container, err := as.Run(ctx, "aerospike/aerospike-server:latest",
			testcontainers.WithEnv(map[string]string{"CUSTOM_ENV": customEnv}))
		require.NoError(t, err)
		require.NotNil(t, container)
		defer container.Container.Terminate(ctx)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := as.Run(ctx, "aerospike/aerospike-server:latest")
		require.Error(t, err)
	})
}
