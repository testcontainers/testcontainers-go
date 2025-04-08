package toxiproxy_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/toxiproxy"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

func TestRun_withPortRange(t *testing.T) {
	ctx := context.Background()

	t.Run("no-port-range", func(t *testing.T) {
		portsCount := 31 // default port range is 31 (not exposed)

		ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)
		jsonInspect, err := ctr.Inspect(ctx)
		require.NoError(t, err)
		require.Equal(t, portsCount+1, len(jsonInspect.HostConfig.PortBindings))
	})

	t.Run("negative-port", func(t *testing.T) {
		portsCount := -1

		ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", toxiproxy.WithPortRange(portsCount))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
		require.Nil(t, ctr)
	})

	t.Run("zero-port", func(t *testing.T) {
		portsCount := 0

		ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", toxiproxy.WithPortRange(portsCount))
		testcontainers.CleanupContainer(t, ctr)
		require.Error(t, err)
		require.Nil(t, ctr)
	})

	t.Run("one-port", func(t *testing.T) {
		portsCount := 1

		ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", toxiproxy.WithPortRange(portsCount))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		jsonInspect, err := ctr.Inspect(ctx)
		require.NoError(t, err)
		require.Equal(t, portsCount+1, len(jsonInspect.HostConfig.PortBindings))
	})

	t.Run("more-than-default-port", func(t *testing.T) {
		portsCount := 75

		ctr, err := toxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0", toxiproxy.WithPortRange(portsCount))
		testcontainers.CleanupContainer(t, ctr)
		require.NoError(t, err)

		jsonInspect, err := ctr.Inspect(ctx)
		require.NoError(t, err)
		require.Equal(t, portsCount+1, len(jsonInspect.HostConfig.PortBindings))
	})
}
