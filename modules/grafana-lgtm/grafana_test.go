package grafanalgtm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	grafanalgtm "github.com/testcontainers/testcontainers-go/modules/grafana-lgtm"
)

func TestGrafanaLGTM(t *testing.T) {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	testcontainers.CleanupContainer(t, grafanaLgtmContainer)
	require.NoError(t, err)

	// perform assertions

	t.Run("right-version", func(t *testing.T) {
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

		body := make(map[string]any)
		err = json.NewDecoder(httpResp.Body).Decode(&body)
		require.NoError(t, err)
		require.Equal(t, "11.0.0", body["version"])
	})

	availableURL := func(t *testing.T, url string) {
		t.Helper()

		conn, err := net.Dial("tcp", url)
		defer func() {
			if conn != nil {
				err := conn.Close()
				require.NoError(t, err)
			}
		}()
		require.NoError(t, err)
	}

	t.Run("loki-endpoint", func(t *testing.T) {
		lokiEndpoint := grafanaLgtmContainer.MustLokiEndpoint(ctx)
		require.NotEmpty(t, lokiEndpoint)
		availableURL(t, lokiEndpoint)
	})

	t.Run("tempo-endpoint", func(t *testing.T) {
		tempoEndpoint := grafanaLgtmContainer.MustTempoEndpoint(ctx)
		require.NotEmpty(t, tempoEndpoint)
		availableURL(t, tempoEndpoint)
	})

	t.Run("otlp-http-endpoint", func(t *testing.T) {
		otlpHTTPEndpoint := grafanaLgtmContainer.MustOtlpHttpEndpoint(ctx)
		require.NotEmpty(t, otlpHTTPEndpoint)
		availableURL(t, otlpHTTPEndpoint)
	})

	t.Run("otlp-grpc-endpoint", func(t *testing.T) {
		otlpGrpcEndpoint := grafanaLgtmContainer.MustOtlpGrpcEndpoint(ctx)
		require.NotEmpty(t, otlpGrpcEndpoint)
		availableURL(t, otlpGrpcEndpoint)
	})

	t.Run("prometheus-http-endpoint", func(t *testing.T) {
		prometheusHTTPEndpoint := grafanaLgtmContainer.MustPrometheusHttpEndpoint(ctx)
		require.NotEmpty(t, prometheusHTTPEndpoint)
		availableURL(t, prometheusHTTPEndpoint)
	})

	t.Run("grafana-endpoint", func(t *testing.T) {
		grafanaEndpoint := grafanaLgtmContainer.MustHttpEndpoint(ctx)
		require.NotEmpty(t, grafanaEndpoint)
		availableURL(t, grafanaEndpoint)
	})
}
