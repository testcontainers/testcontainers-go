package redis_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7")
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithConfigFile(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "/usr/local/redis.conf")

	assertSetsGets(t, ctx, redisContainer, 1)
}

func TestRedisWithLogLevel(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithLogLevel(tcredis.LogLevelVerbose))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "--loglevel", "verbose")

	assertSetsGets(t, ctx, redisContainer, 10)
}

func TestRedisWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithSnapshotting(10, 1))
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err)

	jsonInspect, err := redisContainer.Inspect(ctx)
	require.NoError(t, err)

	require.Contains(t, jsonInspect.Config.Cmd, "--save", "10", "1")

	assertSetsGets(t, ctx, redisContainer, 10)
}

func TestRedisWithTLS(t *testing.T) {
	ctx := context.Background()

	t.Run("mtls-enabled", func(t *testing.T) {
		redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithTLS())
		testcontainers.CleanupContainer(t, redisContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, redisContainer, 1)
	})

	t.Run("mtls-disabled", func(t *testing.T) {
		redisContainer, err := tcredis.Run(ctx, "redis:7", tcredis.WithTLS(), testcontainers.WithCmdArgs("--tls-auth-clients", "no"))
		testcontainers.CleanupContainer(t, redisContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, redisContainer, 1)
	})
}

func assertSetsGets(t *testing.T, ctx context.Context, redisContainer *tcredis.RedisContainer, keyCount int) {
	t.Helper()

	uri, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	endpoint, err := url.Parse(uri)
	require.NoError(t, err)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	if config := redisContainer.TLSConfig(); config != nil {
		require.Equal(t, "rediss", endpoint.Scheme)
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: config}).DialContext(ctx, "tcp", endpoint.Host)
	} else {
		require.Equal(t, "redis", endpoint.Scheme)
		conn, err = dialer.DialContext(ctx, "tcp", endpoint.Host)
	}
	require.NoError(t, err)
	reader := bufio.NewReader(conn)
	t.Cleanup(func() {
		defer func() { require.NoError(t, conn.Close()) }()
		require.Equal(t, "OK", redisCommand(t, conn, reader, "FLUSHALL"))
	})

	t.Log("pinging redis")
	pong := redisCommand(t, conn, reader, "PING")
	t.Log("received response from redis")
	require.Equalf(t, "PONG", pong, "received unexpected response from redis: %s", pong)

	for i := 0; i < keyCount; i++ {
		key := fmt.Sprintf("{user.%s}.favoritefood.%d", uuid.NewString(), i)
		value := fmt.Sprintf("Cabbage Biscuits %d", i)
		require.Equal(t, "OK", redisCommand(t, conn, reader, "SET", key, value, "EX", "7200"))
		savedValue := redisCommand(t, conn, reader, "GET", key)
		require.Equal(t, value, savedValue)
	}
}

// redisCommand handles only the simple and bulk string replies used by these tests.
func redisCommand(t *testing.T, conn net.Conn, reader *bufio.Reader, args ...string) string {
	t.Helper()
	require.NoError(t, conn.SetDeadline(time.Now().Add(10*time.Second)))

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
