package weaviate_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/weaviate"
)

func TestWeaviate(t *testing.T) {
	ctx := context.Background()

	ctr, err := weaviate.Run(ctx, "semitechnologies/weaviate:1.29.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("HttpHostAddress", func(t *testing.T) {
		// httpHostAddress {
		schema, host, err := ctr.HttpHostAddress(ctx)
		// }
		require.NoError(t, err)

		cli := &http.Client{}
		resp, err := cli.Get(fmt.Sprintf("%s://%s", schema, host))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GrpcHostAddress", func(t *testing.T) {
		// gRPCHostAddress {
		host, err := ctr.GrpcHostAddress(ctx)
		// }
		require.NoError(t, err)

		endpointHost, endpointPort, err := net.SplitHostPort(host)
		require.NoError(t, err)
		containerHost, err := ctr.Host(ctx)
		require.NoError(t, err)
		mappedPort, err := ctr.MappedPort(ctx, "50051/tcp")
		require.NoError(t, err)
		require.Equal(t, containerHost, endpointHost)
		require.Equal(t, mappedPort.Port(), endpointPort)
	})

	t.Run("Meta endpoint", func(tt *testing.T) {
		httpScheme, httpHost, err := ctr.HttpHostAddress(ctx)
		require.NoError(tt, err)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s://%s/v1/meta", httpScheme, httpHost), nil)
		require.NoError(tt, err)
		resp, err := (&http.Client{}).Do(req)
		require.NoError(tt, err)
		defer resp.Body.Close()
		require.Equal(tt, http.StatusOK, resp.StatusCode)

		var meta struct {
			Version string `json:"version"`
		}
		require.NoError(tt, json.NewDecoder(resp.Body).Decode(&meta))

		require.NotEmptyf(tt, meta.Version, "failed to get /v1/meta response")
	})
}
