package mongotest

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRelayLifetime(t *testing.T) {
	for _, shutdown := range []string{"cleanup", "context cancellation", "upstream EOF", "client EOF"} {
		t.Run(shutdown, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			destination, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			defer destination.Close()
			r, err := newRelay(ctx, destination.Addr().String())
			require.NoError(t, err)
			defer r.close()
			client, err := net.DialTimeout("tcp", r.listener.Addr().String(), time.Second)
			require.NoError(t, err)
			defer client.Close()
			require.NoError(t, destination.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second)))
			upstream, err := destination.Accept()
			require.NoError(t, err)
			defer upstream.Close()
			require.NoError(t, client.SetDeadline(time.Now().Add(2*time.Second)))
			require.NoError(t, upstream.SetDeadline(time.Now().Add(2*time.Second)))

			// Exact destination and bidirectional bytes, without inspecting a protocol.
			_, err = client.Write([]byte("to generated endpoint"))
			require.NoError(t, err)
			data := make([]byte, len("to generated endpoint"))
			_, err = io.ReadFull(upstream, data)
			require.NoError(t, err)
			require.Equal(t, "to generated endpoint", string(data))
			_, err = upstream.Write([]byte("from generated endpoint"))
			require.NoError(t, err)
			data = make([]byte, len("from generated endpoint"))
			_, err = io.ReadFull(client, data)
			require.NoError(t, err)
			require.Equal(t, "from generated endpoint", string(data))

			switch shutdown {
			case "cleanup":
				r.close()
			case "context cancellation":
				cancel()
				select {
				case <-r.done:
				case <-time.After(time.Second):
					t.Fatal("relay workers did not finish")
				}
			case "upstream EOF":
				require.NoError(t, upstream.Close())
			case "client EOF":
				require.NoError(t, client.Close())
			}
			if shutdown != "client EOF" {
				_, err = client.Read(make([]byte, 1))
				require.Error(t, err)
				require.NotErrorIs(t, err, context.DeadlineExceeded)
				var e net.Error
				if errors.As(err, &e) {
					require.False(t, e.Timeout(), "relay failed to close client")
				}
			}
			if shutdown != "upstream EOF" {
				_, err = upstream.Read(make([]byte, 1))
				require.Error(t, err)
				var e net.Error
				if errors.As(err, &e) {
					require.False(t, e.Timeout(), "relay failed to close upstream")
				}
			}
			r.close() // Idempotent cleanup also waits for both copy directions.
			r.mu.Lock()
			require.Empty(t, r.sockets)
			r.mu.Unlock()
			_, err = net.DialTimeout("tcp", r.listener.Addr().String(), time.Second)
			require.Error(t, err, "listener survived cleanup")
		})
	}
}

func TestRelayDialFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Port zero cannot be a listening service; failed outbound dials must close the
	// accepted socket, rather than keeping mongosh connected to an idle relay.
	r, err := newRelay(ctx, "127.0.0.1:0")
	require.NoError(t, err)
	defer r.close()
	client, err := net.DialTimeout("tcp", r.listener.Addr().String(), time.Second)
	require.NoError(t, err)
	defer client.Close()
	require.NoError(t, client.SetReadDeadline(time.Now().Add(time.Second)))
	_, err = client.Read(make([]byte, 1))
	require.Error(t, err)
	var e net.Error
	if errors.As(err, &e) {
		require.False(t, e.Timeout())
	}
	r.close()
	r.mu.Lock()
	require.Empty(t, r.sockets)
	r.mu.Unlock()
}
