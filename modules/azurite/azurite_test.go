package azurite_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
)

func TestAzurite(t *testing.T) {
	ctx := t.Context()

	ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.23.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
