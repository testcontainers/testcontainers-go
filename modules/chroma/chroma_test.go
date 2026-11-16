package chroma_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func TestChroma(t *testing.T) {
	ctx := context.Background()

	ctr, err := chroma.Run(ctx, "chromadb/chroma:1.4.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("REST Endpoint retrieve docs site", func(tt *testing.T) {
		// restEndpoint {
		restEndpoint, err := ctr.RESTEndpoint(ctx)
		// }
		require.NoErrorf(tt, err, "failed to get REST endpoint")

		cli := &http.Client{}
		resp, err := cli.Get(restEndpoint + "/docs")
		require.NoErrorf(tt, err, "failed to perform GET request")
		defer resp.Body.Close()

		require.Equalf(tt, http.StatusOK, resp.StatusCode, "unexpected status code: %d", resp.StatusCode)
	})

	t.Run("Heartbeat", func(tt *testing.T) {
		endpoint, err := ctr.RESTEndpoint(context.Background())
		require.NoErrorf(tt, err, "failed to get REST endpoint")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"/api/v2/heartbeat", nil)
		require.NoError(tt, err)
		resp, err := (&http.Client{}).Do(req)
		require.NoError(tt, err)
		defer resp.Body.Close()
		require.Equal(tt, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(tt, err)
		require.Contains(tt, string(body), "nanosecond heartbeat")
	})
}
