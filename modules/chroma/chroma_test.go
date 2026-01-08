package chroma_test

import (
	"net/http"
	"testing"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func TestChroma(t *testing.T) {
	ctr, err := chroma.Run(t.Context(), "chromadb/chroma:0.4.24")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("REST Endpoint retrieve docs site", func(tt *testing.T) {
		ctx := tt.Context()
		// restEndpoint {
		restEndpoint, err := ctr.RESTEndpoint(ctx)
		// }
		require.NoErrorf(tt, err, "failed to get REST endpoint")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, restEndpoint+"/docs", http.NoBody)
		require.NoErrorf(tt, err, "failed to create request")

		resp, err := http.DefaultClient.Do(req)
		require.NoErrorf(tt, err, "failed to perform GET request")

		require.Equalf(tt, http.StatusOK, resp.StatusCode, "unexpected status code: %d", resp.StatusCode)
		require.NoError(tt, resp.Body.Close())
	})

	t.Run("GetClient", func(tt *testing.T) {
		ctx := tt.Context()
		// restEndpoint {
		endpoint, err := ctr.RESTEndpoint(ctx)
		require.NoErrorf(tt, err, "failed to get REST endpoint")
		chromaClient, err := chromago.NewClient(endpoint)
		// }
		require.NoErrorf(tt, err, "failed to create client")

		hb, err := chromaClient.Heartbeat(ctx)
		require.NoError(tt, err)
		require.NotNil(tt, hb["nanosecond heartbeat"])
	})
}
