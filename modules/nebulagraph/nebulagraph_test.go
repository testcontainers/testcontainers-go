package nebulagraph_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nebulagraph"
)

const (
	defaultGraphdImage   = "vesoft/nebula-graphd:v3.8.0"
	defaultMetadImage    = "vesoft/nebula-metad:v3.8.0"
	defaultStoragedImage = "vesoft/nebula-storaged:v3.8.0"
)

func TestNebulaGraphContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	container, err := nebulagraph.RunCluster(ctx,
		defaultGraphdImage, []testcontainers.ContainerCustomizer{},
		defaultStoragedImage, []testcontainers.ContainerCustomizer{},
		defaultMetadImage, []testcontainers.ContainerCustomizer{},
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	conn, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, conn)

	// Parse the connection string to get host and port
	host, portt, err := net.SplitHostPort(conn)
	require.NoError(t, err)

	portInt, err := strconv.Atoi(portt)
	require.NoError(t, err)

	t.Run("basic-operations", func(t *testing.T) {
		require.NotEmpty(t, host)
		require.Positive(t, portInt)
	})
}
