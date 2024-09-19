package opensearch_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/opensearch"
)

func TestOpenSearch(t *testing.T) {
	ctx := context.Background()

	ctr, err := opensearch.Run(ctx, "opensearchproject/opensearch:2.11.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("Connect to Address", func(t *testing.T) {
		address, err := ctr.Address(ctx)
		require.NoError(t, err)

		client := &http.Client{}

		req, err := http.NewRequest("GET", address, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
	})
}
