package qdrant_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, restEndpoint, http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("gRPC Endpoint", func(t *testing.T) {
		// gRPCEndpoint {
		grpcEndpoint, err := ctr.GRPCEndpoint(ctx)
		// }
		require.NoError(t, err)

		conn, err := grpc.NewClient(grpcEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()
	})

	t.Run("Web UI", func(t *testing.T) {
		// webUIEndpoint {
		webUI, err := ctr.WebUI(ctx)
		// }
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, webUI, http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
