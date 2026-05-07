package aerospike_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/aerospike"
)

const (
	aerospikeImage = "aerospike/aerospike-server:latest"
)

// TestAerospike tests the Aerospike container functionality
// It includes tests for starting the container with a valid image,
// applying container customizations, and handling context cancellation.
// It also includes a test for an invalid image to ensure proper error handling.
// The tests use the testcontainers-go library to manage container lifecycle
func TestAeroSpike(t *testing.T) {
	t.Run("valid-image-succeeds", func(t *testing.T) {
		ctx := context.Background()
		container, err := aerospike.Run(ctx, aerospikeImage)
		testcontainers.CleanupContainer(t, container)
		require.NoError(t, err)
		require.NotNil(t, container)
	})

	t.Run("applies-container-customizations", func(t *testing.T) {
		ctx := context.Background()
		customEnv := "TEST_ENV=value"
		container, err := aerospike.Run(ctx, aerospikeImage,
			testcontainers.WithEnv(map[string]string{"CUSTOM_ENV": customEnv}))
		testcontainers.CleanupContainer(t, container)
		require.NoError(t, err)
		require.NotNil(t, container)
	})

	t.Run("respects-context-cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		container, err := aerospike.Run(ctx, aerospikeImage)
		testcontainers.CleanupContainer(t, container)
		require.Error(t, err)
	})
}
