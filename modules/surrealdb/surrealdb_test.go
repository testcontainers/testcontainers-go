package surrealdb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestSurrealDBSelect(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "surrealdb/surrealdb:v1.1.1")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	url, err := ctr.URL(ctx)
	require.NoError(t, err)

	assertHTTPDocument(t, ctx, url, false)
}

func TestSurrealDBWithAuth(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "surrealdb/surrealdb:v1.1.1", WithAuthentication())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// websocketURL {
	url, err := ctr.URL(ctx)
	// }
	require.NoError(t, err)

	assertHTTPDocument(t, ctx, url, true)
}

func TestSurrealDBWithAllowAllCaps(t *testing.T) {
	ctx := context.Background()

	ctr, err := Run(ctx, "surrealdb/surrealdb:v1.1.1", WithAllowAllCaps())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	inspect, err := ctr.Inspect(ctx)
	require.NoError(t, err)

	for _, env := range inspect.Config.Env {
		if v, ok := strings.CutPrefix(env, "SURREAL_CAPS_ALLOW_ALL="); ok {
			require.Equal(t, "true", v)
			break
		}
	}
}

// The HTTP endpoint is derived from the public WebSocket endpoint; the child
// examples module separately retains the original WebSocket SDK operations.
func assertHTTPDocument(t *testing.T, ctx context.Context, websocketURL string, authenticated bool) {
	t.Helper()
	endpoint, err := url.Parse(websocketURL)
	require.NoError(t, err)
	require.Equal(t, "ws", endpoint.Scheme)
	require.Equal(t, "/rpc", endpoint.Path)
	require.NotEmpty(t, endpoint.Hostname())
	require.NotEmpty(t, endpoint.Port())
	endpoint.Scheme = "http"
	endpoint.Path = "/sql"
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(`
CREATE person:tobie CONTENT {title: "Founder & CEO", name: {first: "Tobie", last: "Morgan Hitchcock"}, marketing: true};
SELECT * FROM person:tobie;
`))
	require.NoError(t, err)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("NS", "test")
	req.Header.Set("DB", "test")
	if authenticated {
		req.SetBasicAuth("root", "root")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	const maxResponse = 1 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponse+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(body), maxResponse)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", body)
	var statements []struct {
		Status string          `json:"status"`
		Result json.RawMessage `json:"result"`
	}
	require.NoError(t, json.Unmarshal(body, &statements))
	require.Len(t, statements, 2)
	for _, statement := range statements {
		require.Equal(t, "OK", statement.Status, "%s", statement.Result)
	}
	var result []map[string]any
	require.NoError(t, json.Unmarshal(statements[1].Result, &result))
	require.Len(t, result, 1)
	resultData := result[0]
	require.Equal(t, "Founder & CEO", resultData["title"])
	require.Equal(t, "Tobie", resultData["name"].(map[string]any)["first"])
	require.Equal(t, "Morgan Hitchcock", resultData["name"].(map[string]any)["last"])
	require.Equal(t, true, resultData["marketing"])
}
