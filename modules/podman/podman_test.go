package podman

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestRegistration(t *testing.T) {
	require.True(t, core.HasRegisteredProvider("podman"))
	t.Setenv("TESTCONTAINERS_PROVIDER", "podman")
	cfg, err := core.Runtime(t.Context())
	require.NoError(t, err)
	require.Equal(t, "podman", cfg.Name)
	require.Equal(t, "podman", cfg.DefaultBridgeNetwork)
}

func TestInitialize(t *testing.T) {
	cfg, err := (&Provider{}).Initialize(t.Context())
	require.NoError(t, err)
	require.Equal(t, "podman", cfg.DefaultBridgeNetwork)
	require.Empty(t, cfg.DockerHost)
	require.Empty(t, cfg.DockerSocketPath)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = (&Provider{}).Initialize(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
