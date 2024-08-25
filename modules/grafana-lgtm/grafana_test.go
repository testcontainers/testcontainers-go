package grafanalgtm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/grafanalgtm"
)

func TestGrafanaLGTM(t *testing.T) {
	ctx := context.Background()

	grafanaLgtmContainer, err := grafanalgtm.Run(ctx, "grafana/otel-lgtm:0.6.0")
	if err != nil {
		t.Fatal(err)
	}

	// Clean up the container after the test is complete
	t.Cleanup(func() {
		if err := grafanaLgtmContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	// perform assertions

	t.Run("container is running with right version", func(t *testing.T) {
		healthURL, err := url.Parse(fmt.Sprintf("http://%s/api/health", grafanaLgtmContainer.MustHttpEndpoint(ctx)))
		if err != nil {
			t.Fatal(err)
		}

		httpReq := http.Request{
			Method: http.MethodGet,
			URL:    healthURL,
		}

		httpClient := http.Client{}

		httpResp, err := httpClient.Do(&httpReq)
		if err != nil {
			t.Fatal(err)
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d, got %d", http.StatusOK, httpResp.StatusCode)
		}

		body := make(map[string]interface{})
		err = json.NewDecoder(httpResp.Body).Decode(&body)
		if err != nil {
			t.Fatal(err)
		}

		if body["version"] != "11.0.0" {
			t.Fatalf("expected version %q, got %q", "11.0.0", body["version"])
		}
	})
}
