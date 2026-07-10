package timeplus_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/timeplus"
)

func TestTimeplus(t *testing.T) {
	ctx := context.Background()

	ctr, err := timeplus.Run(ctx, "timeplus/timeplusd:2.3")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	t.Run("http endpoint", func(t *testing.T) {
		endpoint, err := ctr.HTTPEndpoint(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, endpoint)

		resp, err := http.Get(endpoint + "/ping") //nolint:noctx
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
