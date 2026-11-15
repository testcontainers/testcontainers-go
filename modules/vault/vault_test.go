package vault_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/testcontainers/testcontainers-go"
	testcontainervault "github.com/testcontainers/testcontainers-go/modules/vault"
)

const (
	token = "root-token"
)

func TestVault(t *testing.T) {
	ctx := context.Background()
	opts := []testcontainers.ContainerCustomizer{
		// WithToken {
		testcontainervault.WithToken(token),
		// }
		// WithInitCommand {
		testcontainervault.WithInitCommand("secrets enable transit", "write -f transit/keys/my-key"),
		testcontainervault.WithInitCommand("kv put secret/test1 foo1=bar1"),
		// }
	}

	vaultContainer, err := testcontainervault.Run(ctx, "hashicorp/vault:1.13.0", opts...)
	testcontainers.CleanupContainer(t, vaultContainer)
	require.NoError(t, err)

	// httpHostAddress {
	hostAddress, err := vaultContainer.HttpHostAddress(ctx)
	// }
	require.NoError(t, err)

	t.Run("Get secret path", func(t *testing.T) {
		t.Run("From vault CLI", func(t *testing.T) {
			ctx := context.Background()

			// containerCliRead {
			exec, reader, err := vaultContainer.Exec(ctx, []string{"vault", "kv", "get", "-format=json", "secret/test1"})
			require.NoError(t, err)
			require.Zero(t, exec)

			bytes, err := io.ReadAll(reader)
			require.NoError(t, err)

			assert.Equal(t, "bar1", gjson.Get(string(bytes), "data.data.foo1").String())
			// }
		})

		t.Run("From HTTP request", func(t *testing.T) {
			// httpRead {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, hostAddress+"/v1/secret/data/test1", nil)
			require.NoError(t, err)
			request.Header.Set("X-Vault-Token", token)

			client := &http.Client{Timeout: 30 * time.Second}
			response, err := client.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()

			const maxBody = 1 << 20
			body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
			require.NoError(t, err)
			require.LessOrEqual(t, len(body), maxBody, "Vault response exceeds limit")
			require.Equal(t, http.StatusOK, response.StatusCode, "%s", body)

			assert.Equal(t, "bar1", gjson.Get(string(body), "data.data.foo1").String())
			// }
		})

		t.Run("From KV-v2 HTTP request", func(t *testing.T) {
			data := kvV2Request(t, ctx, hostAddress, http.MethodGet, "test1", nil)
			assert.Equal(t, "bar1", data["foo1"])
		})
	})

	t.Run("Write secret", func(t *testing.T) {
		t.Run("From KV-v2 HTTP request", func(t *testing.T) {
			kvV2Request(t, ctx, hostAddress, http.MethodPost, "test3", strings.NewReader(`{"data":{"foo":"bar"}}`))

			data := kvV2Request(t, ctx, hostAddress, http.MethodGet, "test3", nil)
			assert.Equal(t, "bar", data["foo"])
		})
	})
}

// kvV2Request exercises the KV-v2 data API used by vault-client-go v0.4.3.
func kvV2Request(t *testing.T, ctx context.Context, hostAddress, method, path string, body io.Reader) map[string]any {
	t.Helper()
	request, err := http.NewRequestWithContext(ctx, method, hostAddress+"/v1/secret/data/"+path, body)
	require.NoError(t, err)
	request.Header.Set("X-Vault-Token", token)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	const maxBody = 1 << 20
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(responseBody), maxBody, "Vault response exceeds limit")
	require.Equal(t, http.StatusOK, response.StatusCode, "%s", responseBody)

	var result struct {
		Data struct {
			Data map[string]any `json:"data"`
		} `json:"data"`
		Errors []string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(responseBody, &result))
	require.Empty(t, result.Errors)
	return result.Data.Data
}
