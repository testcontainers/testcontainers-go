package valkey_test

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
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

func TestIntegrationSetGet(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5")
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithConfigFile(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5", tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 1)
}

func TestValkeyWithImage(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		image string
	}{
		// There is only one release of Valkey at the time of writing
		{
			name:  "Valkey7.2.5",
			image: "valkey/valkey:7.2.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valkeyContainer, err := tcvalkey.Run(ctx, tt.image, tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")))
			testcontainers.CleanupContainer(t, valkeyContainer)
			require.NoError(t, err)

			assertSetsGets(t, ctx, valkeyContainer, 1)
		})
	}
}

func TestValkeyWithLogLevel(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5", tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func TestValkeyWithSnapshotting(t *testing.T) {
	ctx := context.Background()

	valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5", tcvalkey.WithSnapshotting(10, 1))
	testcontainers.CleanupContainer(t, valkeyContainer)
	require.NoError(t, err)

	assertSetsGets(t, ctx, valkeyContainer, 10)
}

func TestRedisWithTLS(t *testing.T) {
	ctx := context.Background()

	t.Run("mtls-disabled", func(t *testing.T) {
		valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5", tcvalkey.WithTLS())
		testcontainers.CleanupContainer(t, valkeyContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, valkeyContainer, 1)
	})

	t.Run("mtls-enabled", func(t *testing.T) {
		valkeyContainer, err := tcvalkey.Run(ctx, "valkey/valkey:7.2.5", tcvalkey.WithTLS(), testcontainers.WithCmdArgs("--tls-auth-clients", "no"))
		testcontainers.CleanupContainer(t, valkeyContainer)
		require.NoError(t, err)

		assertSetsGets(t, ctx, valkeyContainer, 1)
	})
}

func assertSetsGets(t *testing.T, ctx context.Context, valkeyContainer *tcvalkey.ValkeyContainer, keyCount int) {
	t.Helper()
	uri, err := valkeyContainer.ConnectionString(ctx)
	require.NoError(t, err)
	endpoint, err := url.Parse(uri)
	require.NoError(t, err)

	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	if config := valkeyContainer.TLSConfig(); config != nil {
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
		require.Equal(t, "OK", valkeyCommand(t, conn, reader, "FLUSHALL"))
	})

	t.Log("pinging valkey")
	pong := valkeyCommand(t, conn, reader, "PING")
	t.Log("received response from valkey")
	require.Equalf(t, "PONG", pong, "received unexpected response from valkey: %s", pong)

	for i := 0; i < keyCount; i++ {
		key := fmt.Sprintf("{user.%s}.favoritefood.%d", uuid.NewString(), i)
		value := fmt.Sprintf("Cabbage Biscuits %d", i)
		ttl := 2 * time.Hour

		require.Equal(t, "OK", valkeyCommand(t, conn, reader, "SET", key, value, "EXAT", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10)))
		require.Equal(t, int64(1), valkeyCommand(t, conn, reader, "EXPIRE", key, strconv.FormatInt(int64(ttl.Seconds()), 10)))
		require.Equal(t, value, valkeyCommand(t, conn, reader, "GET", key))
	}
}

// valkeyCommand handles only the simple string, integer and bulk string replies
// used by these tests, with bounded response sizes and read/write deadlines.
func valkeyCommand(t *testing.T, conn net.Conn, reader *bufio.Reader, args ...string) any {
	t.Helper()
	require.NoError(t, conn.SetDeadline(time.Now().Add(10*time.Second)))

	var request strings.Builder
	fmt.Fprintf(&request, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&request, "$%d\r\n%s\r\n", len(arg), arg)
	}
	_, err := io.WriteString(conn, request.String())
	require.NoError(t, err)
	line, err := reader.ReadSlice('\n')
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(line), 3, "invalid Valkey response: %q", line)
	require.True(t, strings.HasSuffix(string(line), "\r\n"), "invalid Valkey response: %q", line)
	payload := string(line[1 : len(line)-2])
	switch line[0] {
	case '+':
		return payload
	case ':':
		value, err := strconv.ParseInt(payload, 10, 64)
		require.NoError(t, err)
		return value
	case '$':
		size, err := strconv.Atoi(payload)
		require.NoError(t, err)
		require.GreaterOrEqual(t, size, 0, "unexpected nil bulk string")
		require.LessOrEqual(t, size, 1<<20, "unexpected large bulk string")
		data := make([]byte, size+2)
		_, err = io.ReadFull(reader, data)
		require.NoError(t, err)
		require.Equal(t, "\r\n", string(data[size:]))
		return string(data[:size])
	default:
		t.Fatalf("unexpected Valkey response: %q", line)
		return nil
	}
}
