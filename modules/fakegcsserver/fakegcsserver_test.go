// Package fakegcsserver_test — see package documentation in examples_test.go.
package fakegcsserver_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/fakegcsserver"
)

// TestFakeGCSServer verifies that the default (http) container starts correctly,
// returns a valid StorageURL, and responds to bucket-list requests.
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

// TestFakeGCSServerWithHTTPScheme verifies that an explicit http scheme starts
// correctly and that StorageURL returns an http:// URL.
func TestFakeGCSServerWithHTTPScheme(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("http"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.StorageURL(ctx)
	require.NoError(t, err)
	require.Contains(t, url, "http://")
	require.Contains(t, url, "/storage/v1")
}

// TestFakeGCSServerWithHTTPSScheme verifies that the https scheme starts
// correctly and that StorageURL returns an https:// URL.
func TestFakeGCSServerWithHTTPSScheme(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("https"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.StorageURL(ctx)
	require.NoError(t, err)
	require.Contains(t, url, "https://")
	require.Contains(t, url, "/storage/v1")
}

// TestFakeGCSServerWithBothSchemeRejected verifies that the unsupported "both"
// scheme is rejected at option-validation time with no container started.
func TestFakeGCSServerWithBothSchemeRejected(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("both"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.Nil(t, ctr)
}

// TestFakeGCSServerWithInvalidScheme verifies that passing an unsupported
// scheme returns an error and a nil container.
func TestFakeGCSServerWithInvalidScheme(t *testing.T) {
	ctx := context.Background()

	ctr, err := fakegcsserver.Run(ctx, "fsouza/fake-gcs-server:1.47",
		fakegcsserver.WithScheme("ftp"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.Error(t, err)
	require.Nil(t, ctr)
}
