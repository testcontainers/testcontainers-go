package fakegcsserver_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/fakegcsserver"
)

func TestFakeGCSServer(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("StorageURL", func(t *testing.T) {
		url, err := ctr.StorageURL(ctx)
		require.NoError(t, err)
		require.Contains(t, url, "http://")
		require.Contains(t, url, "/storage/v1")
	})

	t.Run("ListBuckets", func(t *testing.T) {
		url, err := ctr.StorageURL(ctx)
		require.NoError(t, err)

		resp, err := http.Get(url + "/b?project=test") //nolint:noctx
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Less(t, resp.StatusCode, 500)
	})
}

func TestFakeGCSServerWithScheme(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("http"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.StorageURL(ctx)
	require.NoError(t, err)
	require.Contains(t, url, "http://")
}

func TestFakeGCSServerWithInvalidScheme(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("ftp"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.Nil(t, ctr)
}
