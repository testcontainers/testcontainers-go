package openfga_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func TestOpenFGA(t *testing.T) {
	ctx := context.Background()

	ctr, err := openfga.Run(ctx, "openfga/openfga:v1.5.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
