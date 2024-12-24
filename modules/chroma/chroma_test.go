package chroma_test

import (
	"context"
	"net/http"
	"testing"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/chroma"
)

func TestChroma(t *testing.T) {
	ctx := context.Background()

	ctr, err := chroma.Run(ctx, "chromadb/chroma:0.4.24")
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

	t.Run("GetClient", func(tt *testing.T) {
		// restEndpoint {
		endpoint, err := ctr.RESTEndpoint(context.Background())
		require.NoErrorf(tt, err, "failed to get REST endpoint")
		chromaClient, err := chromago.NewClient(endpoint)
		// }
		require.NoErrorf(tt, err, "failed to create client")

		hb, err := chromaClient.Heartbeat(context.TODO())
		require.NoError(tt, err)
		require.NotNil(tt, hb["nanosecond heartbeat"])
	})
}
