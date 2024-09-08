package wait_test

import (
	"net"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/wait"
)

// testNatPort creates a new NAT port for testing.
func testNatPort(t *testing.T) nat.Port {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, listener.Close())
	})

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	natPort, err := nat.NewPort("tcp", port)
	require.NoError(t, err)

	return natPort
}

// testPortStrategy tests the given wait.Strategy with the given option.
func testHostPortStrategy(t *testing.T, strategy wait.Strategy) {
	t.Helper()

	testPortScenarios(t, strategy, func(t *testing.T, b *waitBuilder) *waitBuilder {
		if b.NoTCP() {
			// No TCP ports so listener needed.
			return b
		}

		port := testNatPort(t)
		return b.MappedPorts(port).Exec(1, 0)
	})

	t.Run("no-shell", func(t *testing.T) {
		port := testNatPort(t)
		newWaitBuilder().
			MappedPorts(port).
			Exec(126).
			Run(t, strategy)
	})

	t.Run("no-port", func(t *testing.T) {
		strategy := &wait.HostPortStrategy{}
		strategy = strategy.
			WithStartupTimeout(200 * time.Millisecond).
			WithPollInterval(10 * time.Millisecond)
		var portErr wait.PortNotFoundErr
		newWaitBuilder().
			MappedPorts().
			ErrorAs(&portErr).
			Run(t, strategy)
	})
}

func TestForListeningPort(t *testing.T) {
	strategy := wait.ForListeningPort("80").
		WithStartupTimeout(200 * time.Millisecond).
		WithPollInterval(10 * time.Millisecond)

	testHostPortStrategy(t, strategy)
}

func TestWaitForExposed(t *testing.T) {
	strategy := wait.ForExposedPort().
		WithStartupTimeout(200 * time.Millisecond).
		WithPollInterval(10 * time.Millisecond)

	testHostPortStrategy(t, strategy)
}

func TestNewHostPortStrategy(t *testing.T) {
	strategy := wait.NewHostPortStrategy("80").
		WithStartupTimeout(200 * time.Millisecond).
		WithPollInterval(10 * time.Millisecond)

	testHostPortStrategy(t, strategy)
}
