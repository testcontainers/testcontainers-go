package toxiproxy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/toxiproxy"
)

func TestToxiproxy(t *testing.T) {
	ctx := context.Background()

	ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}
