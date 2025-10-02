package openfga_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/openfga"
)

func TestOpenFGA(t *testing.T) {
	ctx := t.Context()

	ctr, err := openfga.Run(ctx, "openfga/openfga:v1.5.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
