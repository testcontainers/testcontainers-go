package qdrant_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/qdrant"
)

func TestQdrant(t *testing.T) {
	ctx := context.Background()

	ctr, err := qdrant.Run(ctx, "qdrant/qdrant:v1.7.4")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("REST Endpoint", func(t *testing.T) {
		// restEndpoint {
		restEndpoint, err := ctr.RESTEndpoint(ctx)
		// }
		require.NoError(t, err)

		cli := &http.Client{}
		resp, err := cli.Get(restEndpoint)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("gRPC Endpoint", func(t *testing.T) {
		// gRPCEndpoint {
		grpcEndpoint, err := ctr.GRPCEndpoint(ctx)
		// }
		require.NoError(t, err)

		dialer := net.Dialer{Timeout: 5 * time.Second}
		conn, err := dialer.DialContext(ctx, "tcp", grpcEndpoint)
		require.NoError(t, err)
		defer conn.Close()
	})

	t.Run("Web UI", func(t *testing.T) {
		// webUIEndpoint {
		webUI, err := ctr.WebUI(ctx)
		// }
		require.NoError(t, err)

		cli := &http.Client{}
		resp, err := cli.Get(webUI)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
