package typesense_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/typesense"
)

func TestTypesense(t *testing.T) {
	ctx := context.Background()

	ctr, err := typesense.Run(ctx, "typesense/typesense:26.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Address", func(t *testing.T) {
		address, err := ctr.Address(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, address)

		resp, err := http.Get(address + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("APIKey", func(t *testing.T) {
		require.Equal(t, "test-api-key", ctr.APIKey())
	})
}

func TestTypesense_WithAPIKey(t *testing.T) {
	ctx := context.Background()

	ctr, err := typesense.Run(
		ctx,
		"typesense/typesense:26.0",
		typesense.WithAPIKey("custom-key"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "custom-key", ctr.APIKey())
}

// TestTypesense_WithEnvOverride verifies that APIKey() reflects the effective
// env var even when the caller overrides TYPESENSE_API_KEY directly via
// testcontainers.WithEnv (bypassing WithAPIKey).
func TestTypesense_WithEnvOverride(t *testing.T) {
	ctx := context.Background()

	ctr, err := typesense.Run(
		ctx,
		"typesense/typesense:26.0",
		testcontainers.WithEnv(map[string]string{
			"TYPESENSE_API_KEY": "env-override-key",
		}),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	require.Equal(t, "env-override-key", ctr.APIKey())
}
