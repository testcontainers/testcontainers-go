package vearch_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vearch"
)

func TestVearch(t *testing.T) {
	ctx := t.Context()

	ctr, err := vearch.Run(ctx, "vearch/vearch:3.5.1")
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
}
