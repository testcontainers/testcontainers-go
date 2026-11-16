package pinecone_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcpinecone "github.com/testcontainers/testcontainers-go/modules/pinecone"
)

const (
	pineconeAPIKey     = "testcontainers-go"
	pineconeAPIVersion = "2024-10"
)

type pineconeIndexList struct {
	Indexes []struct {
		Name string `json:"name"`
	} `json:"indexes"`
}

func TestPinecone(t *testing.T) {
	ctx := context.Background()
	ctr, err := tcpinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:v0.7.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.HttpEndpoint()
	require.NoError(t, err)

	indexes := listIndexes(t, ctx, host)
	require.Empty(t, indexes.Indexes)
}

func TestPinecone_index(t *testing.T) {
	ctx := context.Background()
	ctr, err := tcpinecone.Run(ctx, "ghcr.io/pinecone-io/pinecone-local:v0.7.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	host, err := ctr.HttpEndpoint()
	require.NoError(t, err)

	indexName := "my-serverless-index"
	createIndex(t, ctx, host, indexName)

	indexes := listIndexes(t, ctx, host)
	require.NotEmpty(t, indexes.Indexes)
	require.Equal(t, indexName, indexes.Indexes[0].Name)
}

func listIndexes(t *testing.T, ctx context.Context, host string) pineconeIndexList {
	t.Helper()
	response := pineconeRequest(t, ctx, http.MethodGet, host+"/indexes", nil)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var indexes pineconeIndexList
	require.NoError(t, json.NewDecoder(response.Body).Decode(&indexes))
	return indexes
}

func createIndex(t *testing.T, ctx context.Context, host, indexName string) {
	t.Helper()
	body := map[string]any{
		"name":      indexName,
		"dimension": 3,
		"metric":    "cosine",
		"spec":      map[string]any{"serverless": map[string]string{"cloud": "aws", "region": "us-east-1"}},
		"tags":      map[string]string{"environment": "development"},
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	response := pineconeRequest(t, ctx, http.MethodPost, host+"/indexes", bytes.NewReader(payload))
	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
}

func pineconeRequest(t *testing.T, ctx context.Context, method, endpoint string, body io.Reader) *http.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	require.NoError(t, err)
	req.Header.Set("Api-Key", pineconeAPIKey)
	req.Header.Set("X-Pinecone-Api-Version", pineconeAPIVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("Pinecone request failed: %v", err)
	}
	if response.StatusCode >= http.StatusBadRequest {
		defer response.Body.Close()
		message, _ := io.ReadAll(response.Body)
		t.Fatalf("Pinecone request failed: %s: %s", response.Status, strings.TrimSpace(string(message)))
	}
	if response == nil {
		t.Fatal("Pinecone returned no response")
	}
	return response
}
