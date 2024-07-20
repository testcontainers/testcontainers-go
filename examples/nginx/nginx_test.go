package nginx

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestIntegrationNginxLatestReturn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	nginxC, err := startContainer(ctx)
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	resp, err := http.Get(nginxC.URI)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
