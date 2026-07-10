package nginx_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/nginx"
)

func TestNginx(t *testing.T) {
	ctx := context.Background()

	ctr, err := nginx.Run(ctx, "nginx:1.25")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// httpEndpoint {
	httpEndpoint, err := ctr.HTTPEndpoint(ctx)
	// }
	require.NoError(t, err)
	require.NotEmpty(t, httpEndpoint)

	resp, err := http.Get(httpEndpoint) //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNginxWithConfigFile(t *testing.T) {
	ctx := context.Background()

	ctr, err := nginx.Run(ctx, "nginx:1.25",
		nginx.WithConfigFile("testdata/nginx.conf"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpEndpoint, err := ctr.HTTPEndpoint(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, httpEndpoint)

	resp, err := http.Get(httpEndpoint) //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNginxWithCustomConfig(t *testing.T) {
	ctx := context.Background()

	ctr, err := nginx.Run(ctx, "nginx:1.25",
		nginx.WithCustomConfig("testdata/default.conf"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	httpEndpoint, err := ctr.HTTPEndpoint(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, httpEndpoint)

	resp, err := http.Get(httpEndpoint) //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
