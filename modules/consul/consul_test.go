package consul_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/consul"
)

func TestConsul(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		opts []testcontainers.ContainerCustomizer
	}{
		{
			name: "Default",
			opts: []testcontainers.ContainerCustomizer{},
		},
		{
			name: "WithConfigString",
			opts: []testcontainers.ContainerCustomizer{
				consul.WithConfigString(`{ "server":true }`),
			},
		},
		{
			name: "WithConfigFile",
			opts: []testcontainers.ContainerCustomizer{
				consul.WithConfigFile(filepath.Join("testdata", "config.json")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctr, err := consul.Run(ctx, "hashicorp/consul:1.15", test.opts...)
			testcontainers.CleanupContainer(t, ctr)
			require.NoError(t, err)

			// Check if API is up
			host, err := ctr.ApiEndpoint(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, host)

			client := &http.Client{Timeout: 5 * time.Second}
			t.Cleanup(client.CloseIdleConnections)
			endpoint := "http://" + host
			consulRequest(t, ctx, client, http.MethodGet, endpoint, nil)

			registration, err := json.Marshal(map[string]string{"ID": "abcd", "Name": test.name})
			require.NoError(t, err)

			// Register / Unregister service and verify the agent's service list.
			consulRequest(t, ctx, client, http.MethodPut, endpoint+"/v1/agent/service/register", registration)
			services := consulServices(t, ctx, client, endpoint)
			require.Contains(t, services, "abcd")
			require.Equal(t, "abcd", services["abcd"].ID)
			require.Equal(t, test.name, services["abcd"].Service)

			consulRequest(t, ctx, client, http.MethodPut, endpoint+"/v1/agent/service/deregister/abcd", nil)
			require.NotContains(t, consulServices(t, ctx, client, endpoint), "abcd")
		})
	}
}

func consulRequest(t *testing.T, ctx context.Context, client *http.Client, method, endpoint string, body []byte) []byte {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "%s", responseBody)
	return responseBody
}

type consulService struct {
	ID      string
	Service string
}

func consulServices(t *testing.T, ctx context.Context, client *http.Client, endpoint string) map[string]consulService {
	t.Helper()

	body := consulRequest(t, ctx, client, http.MethodGet, endpoint+"/v1/agent/services", nil)
	var services map[string]consulService
	require.NoError(t, json.Unmarshal(body, &services))
	return services
}
