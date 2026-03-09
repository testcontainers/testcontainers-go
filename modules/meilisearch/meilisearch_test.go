package meilisearch_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/meilisearch"
)

func TestMeilisearch(t *testing.T) {
	ctx := context.Background()

	ctr, err := meilisearch.Run(ctx, "getmeili/meilisearch:v1.10.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	address, err := ctr.Address(ctx)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, address, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
}

func TestMeilisearch_WithDataDump(t *testing.T) {
	ctx := context.Background()

	ctr, err := meilisearch.Run(ctx, "getmeili/meilisearch:v1.10.3",
		meilisearch.WithDumpImport("testdata/movies.dump"),
		meilisearch.WithMasterKey("my-master-key"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	address, err := ctr.Address(ctx)
	require.NoError(t, err)

	client := http.DefaultClient

	req, err := http.NewRequest(http.MethodGet, address, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close() // not closing the body in a defer as it's not used anymore

	require.Equal(t, http.StatusOK, resp.StatusCode)

	path, err := url.JoinPath(address, "/indexes/movies/documents/1212")
	require.NoError(t, err)

	req, err = http.NewRequest(http.MethodGet, path, nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer my-master-key")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Assert the response of that document.
	require.JSONEq(t, `{
  "movie_id": 1212,
  "overview": "When a scientists daughter is kidnapped, American Ninja, attempts to find her, but this time he teams up with a youngster he has trained in the ways of the ninja.",
  "poster": "https://image.tmdb.org/t/p/w1280/iuAQVI4mvjI83wnirpD8GVNRVuY.jpg",
  "release_date": 725846400,
  "title": "American Ninja 5"
}`, string(bodyBytes))
}
