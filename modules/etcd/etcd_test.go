package etcd_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/modules/etcd"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	c, r, err := ctr.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, c)

	output, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(output), "default")
}

func TestRunWithDataDir(t *testing.T) {
	ctx := context.Background()

	ctr, err := etcd.Run(ctx, "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithDataDir())
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	c, r, err := ctr.Exec(ctx, []string{"etcdctl", "member", "list"}, tcexec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, c)

	output, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(output), "default")
}

func TestPutGet(t *testing.T) {
	t.Run("single_node", func(t *testing.T) {
		ctr, err := etcd.Run(context.Background(), "gcr.io/etcd-development/etcd:v3.5.14")
		testPutGet(t, ctr, err, 1)
	})
	t.Run("multiple_nodes", func(t *testing.T) {
		ctr, err := etcd.Run(context.Background(), "gcr.io/etcd-development/etcd:v3.5.14", etcd.WithNodes("etcd-1", "etcd-2", "etcd-3"))
		testPutGet(t, ctr, err, 3)
	})
}

func testPutGet(t *testing.T, ctr *etcd.EtcdContainer, err error, wantEndpoints int) {
	t.Helper()

	testcontainers.CleanupContainer(t, ctr)

	require.NoError(t, err)

	ctx := context.Background()

	clientEndpoints, err := ctr.ClientEndpoints(ctx)
	require.NoError(t, err)

	require.Len(t, clientEndpoints, wantEndpoints)

	client := &http.Client{Timeout: 5 * time.Second}
	t.Cleanup(client.CloseIdleConnections)

	key := base64.StdEncoding.EncodeToString([]byte("sample_key"))
	value := base64.StdEncoding.EncodeToString([]byte("sample_value"))
	etcdGatewayRequest(t, ctx, client, clientEndpoints[0]+"/v3/kv/put", map[string]string{
		"key": key, "value": value,
	})

	// Read through every published client endpoint, including each cluster node.
	for _, endpoint := range clientEndpoints {
		body := etcdGatewayRequest(t, ctx, client, endpoint+"/v3/kv/range", map[string]string{"key": key})
		var resp struct {
			Count string `json:"count"`
			Kvs   []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"kvs"`
		}
		require.NoError(t, json.Unmarshal(body, &resp))
		require.Equal(t, "1", resp.Count)
		require.Len(t, resp.Kvs, 1)
		require.Equal(t, key, resp.Kvs[0].Key)
		decodedValue, err := base64.StdEncoding.DecodeString(resp.Kvs[0].Value)
		require.NoError(t, err)
		require.Equal(t, "sample_value", string(decodedValue))
	}
}

func etcdGatewayRequest(t *testing.T, ctx context.Context, client *http.Client, endpoint string, payload map[string]string) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Run returns before the gateway is ready to accept client requests.
	var responseBody []byte
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		require.NoError(collect, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.NoError(collect, err)
		defer resp.Body.Close()

		responseBody, err = io.ReadAll(resp.Body)
		require.NoError(collect, err)
		require.Equal(collect, http.StatusOK, resp.StatusCode, "%s", responseBody)
	}, 10*time.Second, 100*time.Millisecond)
	return responseBody
}
