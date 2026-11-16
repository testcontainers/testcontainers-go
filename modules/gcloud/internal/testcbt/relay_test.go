package testcbt

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRelayFixedDestinationAndCleanup(t *testing.T) {
	upstream, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { upstream.Close() })
	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := upstream.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	relay, err := NewRelay(upstream.Addr().String())
	require.NoError(t, err)
	t.Cleanup(relay.Close)
	client, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", relay.Port()), time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	require.NoError(t, client.SetDeadline(time.Now().Add(3*time.Second)))

	var server net.Conn
	select {
	case server = <-accepted:
	case <-time.After(3 * time.Second):
		t.Fatal("relay did not dial the fixed destination")
	}
	t.Cleanup(func() { server.Close() })
	require.NoError(t, server.SetDeadline(time.Now().Add(3*time.Second)))

	_, err = client.Write([]byte("request"))
	require.NoError(t, err)
	data := make([]byte, len("request"))
	_, err = io.ReadFull(server, data)
	require.NoError(t, err)
	require.Equal(t, "request", string(data))
	_, err = server.Write([]byte("response"))
	require.NoError(t, err)
	data = make([]byte, len("response"))
	_, err = io.ReadFull(client, data)
	require.NoError(t, err)
	require.Equal(t, "response", string(data))

	closed := make(chan struct{})
	go func() {
		relay.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("relay cleanup left an active copying goroutine")
	}
	_, err = client.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
	_, err = server.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
	relay.Close() // Cleanup must be idempotent.
	_, err = net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", relay.Port()), time.Second)
	require.Error(t, err)
}

func TestRelayUnavailableDestination(t *testing.T) {
	upstream, err := net.Listen("tcp4", "127.0.0.1:0")
	require.NoError(t, err)
	target := upstream.Addr().String()
	require.NoError(t, upstream.Close())
	relay, err := NewRelay(target)
	require.NoError(t, err)
	t.Cleanup(relay.Close)
	client, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", relay.Port()), time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })
	require.NoError(t, client.SetDeadline(time.Now().Add(3*time.Second)))
	_, err = client.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
}
