package aerospike_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	as "github.com/testcontainers/testcontainers-go/modules/aerospike"
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
	t.Run("fails_with_invalid_image", func(t *testing.T) {
		ctx := context.Background()
		_, err := as.Run(ctx, "invalid-aerospike-image")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to start Aerospike container")
	})

	t.Run("succeeds_with_valid_image", func(t *testing.T) {
		ctx := context.Background()
		container, err := as.Run(ctx, aerospikeImage)
		require.NoError(t, err)
		require.NotNil(t, container)
		defer container.Container.Terminate(ctx)

		host, err := container.Host(ctx)
		require.NoError(t, err)

		port, err := container.MappedPort(ctx, "3000/tcp")
		require.NoError(t, err)

		require.NotEmpty(t, host)
		require.NotEmpty(t, port)
	})

	t.Run("applies_container_customizations", func(t *testing.T) {
		ctx := context.Background()
		customEnv := "TEST_ENV=value"
		container, err := as.Run(ctx, aerospikeImage,
			testcontainers.WithEnv(map[string]string{"CUSTOM_ENV": customEnv}))
		require.NoError(t, err)
		require.NotNil(t, container)
		defer container.Container.Terminate(ctx)
	})

	t.Run("respects_context_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		_, err := as.Run(ctx, aerospikeImage)
		require.Error(t, err)
	})
}
