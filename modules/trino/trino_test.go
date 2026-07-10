package trino_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/trino"
)

func TestTrino(t *testing.T) {
	ctx := context.Background()

	ctr, err := trino.Run(ctx, "trinodb/trino:435")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Verify the container exposes a valid connection string.
	connStr, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)
	require.Contains(t, connStr, "http://")
	require.Contains(t, connStr, ":8080")
}
