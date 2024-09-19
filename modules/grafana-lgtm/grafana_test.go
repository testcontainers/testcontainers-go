package grafanalgtm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func TestGrafanaLGTM(t *testing.T) {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	testcontainers.CleanupContainer(t, grafanaLgtmContainer)
	require.NoError(t, err)

	// perform assertions

	t.Run("container is running with right version", func(t *testing.T) {
		healthURL, err := url.Parse(fmt.Sprintf("http://%s/api/health", grafanaLgtmContainer.MustHttpEndpoint(ctx)))
		require.NoError(t, err)

		httpReq := http.Request{
			Method: http.MethodGet,
			URL:    healthURL,
		}

		httpClient := http.Client{}

		httpResp, err := httpClient.Do(&httpReq)
		require.NoError(t, err)

		defer httpResp.Body.Close()

		require.Equal(t, http.StatusOK, httpResp.StatusCode)

		body := make(map[string]interface{})
		err = json.NewDecoder(httpResp.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "11.0.0", body["version"])
	})
}
