package toxiproxy_test

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	tctoxiproxy "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	ctr, err := tctoxiproxy.Run(ctx, "ghcr.io/shopify/toxiproxy:2.12.0")
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// perform assertions
}

//go:embed testdata/toxiproxy.json
var configFile []byte

func TestRun_withConfigFile(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, nw.Remove(ctx))
	})

	redisContainer, err := tcredis.Run(
		ctx,
		"redis:6-alpine",
		network.WithNetwork([]string{"redis"}, nw),
	)
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	toxiproxyContainer, err := tctoxiproxy.Run(
		ctx,
		"ghcr.io/shopify/toxiproxy:2.12.0",
		tctoxiproxy.WithConfigFile(bytes.NewReader(configFile)),
		network.WithNetwork([]string{"toxiproxy"}, nw),
	)
	testcontainers.CleanupContainer(t, toxiproxyContainer)
	require.NoError(t, err)

	toxiURI, err := toxiproxyContainer.URI(ctx)
	require.NoError(t, err)

	toxiproxyProxyPort, err := toxiproxyContainer.MappedPort(ctx, "8666/tcp")
	require.NoError(t, err)

	toxiproxyProxyHostIP, err := toxiproxyContainer.Host(ctx)
	require.NoError(t, err)

	redisAddr := net.JoinHostPort(toxiproxyProxyHostIP, toxiproxyProxyPort.Port())
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", redisAddr)
	require.NoError(t, err)
	reader := bufio.NewReader(conn)
	t.Cleanup(func() {
		require.Equal(t, "OK", redisCommand(t, conn, reader, "FLUSHALL"))
		require.NoError(t, conn.Close())
	})

	key := fmt.Sprintf("{user.%s}.favoritefood", uuid.NewString())
	value := "Cabbage Biscuits"
	require.Equal(t, "OK", redisCommand(t, conn, reader, "SET", key, value, "EX", "7200"))

	const (
		latency = 1_000
		jitter  = 200
	)

	// Add a latency toxic to the proxy
	type toxic struct {
		Name       string         `json:"name"`
		Type       string         `json:"type"`
		Stream     string         `json:"stream,omitempty"`
		Toxicity   float32        `json:"toxicity"`
		Attributes map[string]any `json:"attributes"`
	}
	toxicOptions := toxic{
		Name:     "latency_down",
		Type:     "latency",
		Toxicity: 1.0,
		Stream:   "downstream",
		Attributes: map[string]any{
			"latency": latency,
			"jitter":  jitter,
		},
	}
	requestBody, err := json.Marshal(toxicOptions)
	require.NoError(t, err)
	// redis is the name of the proxy in the config file.
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, toxiURI+"/proxies/redis/toxics", bytes.NewReader(requestBody))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "toxiproxy-cli")
	client := &http.Client{Timeout: 30 * time.Second}
	defer client.CloseIdleConnections()
	response, err := client.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	const maxResponseSize = 1 << 20
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	require.NoError(t, err)
	require.LessOrEqual(t, len(responseBody), maxResponseSize, "toxic response exceeds size limit")
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var apiError struct {
			Message string `json:"error"`
			Status  int    `json:"status"`
		}
		require.NoError(t, json.Unmarshal(responseBody, &apiError))
		t.Fatalf("failed to add toxic to proxy redis: AddToxic: HTTP %d: %s", apiError.Status, apiError.Message)
	}
	var addedToxic toxic
	require.NoError(t, json.Unmarshal(responseBody, &addedToxic))

	start := time.Now()
	// Get data
	savedValue := redisCommand(t, conn, reader, "GET", key)

	duration := time.Since(start)

	t.Logf("Duration: %s\n", duration)

	// The value is retrieved successfully
	require.Equal(t, value, savedValue)

	// Check that latency is within expected range (800ms-1200ms).
	// The latency toxic adds 1000ms (1000ms +/- 200ms jitter).
	minDuration := (latency - jitter) * time.Millisecond
	maxDuration := (latency + jitter) * time.Millisecond
	require.True(t, duration >= minDuration && duration <= maxDuration)
}

func redisCommand(t *testing.T, conn net.Conn, reader *bufio.Reader, args ...string) string {
	t.Helper()
	require.NoError(t, conn.SetDeadline(time.Now().Add(3*time.Second)))
	var request strings.Builder
	fmt.Fprintf(&request, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&request, "$%d\r\n%s\r\n", len(arg), arg)
	}
	_, err := io.WriteString(conn, request.String())
	require.NoError(t, err)
	line, err := reader.ReadString('\n')
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(line, "\r\n"), "invalid Redis response: %q", line)
	switch line[0] {
	case '+':
		return strings.TrimSuffix(line[1:], "\r\n")
	case '$':
		size, err := strconv.Atoi(strings.TrimSuffix(line[1:], "\r\n"))
		require.NoError(t, err)
		require.GreaterOrEqual(t, size, 0, "unexpected nil bulk string")
		data := make([]byte, size+2)
		_, err = io.ReadFull(reader, data)
		require.NoError(t, err)
		require.Equal(t, "\r\n", string(data[size:]))
		return string(data[:size])
	default:
		t.Fatalf("unexpected Redis response: %q", line)
		return ""
	}
}
