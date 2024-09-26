package vearch_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/vearch"
)

func TestVearch(t *testing.T) {
	ctx := context.Background()

	ctr, err := vearch.Run(ctx, "vearch/vearch:3.5.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("REST Endpoint", func(tt *testing.T) {
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
}
