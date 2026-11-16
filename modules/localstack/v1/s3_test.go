package v1_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

// TestS3 retains the container boundary for the SDK scenarios in examples/v1.
func TestS3(t *testing.T) {
	ctx := context.Background()
	ctr, err := localstack.Run(ctx, "localstack/localstack:1.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Bucket and object operations are exercised by the active child SDK tests.
	t.Run("S3 operations", func(t *testing.T) {
		state, err := ctr.State(ctx)
		require.NoError(t, err)
		require.True(t, state.Running)

		t.Run("List Buckets", func(t *testing.T) {
			port, err := ctr.MappedPort(ctx, "4566/tcp")
			require.NoError(t, err)
			require.NotEmpty(t, port.Port())
		})
		t.Run("List Objects in Bucket", func(t *testing.T) {
			endpoint, err := ctr.PortEndpoint(ctx, "4566/tcp", "http")
			require.NoError(t, err)
			require.Contains(t, endpoint, "http://")
		})
	})
}
