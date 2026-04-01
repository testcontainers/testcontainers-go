package testcontainers

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUDPPortBinding tests the fix for the UDP port binding issue.
// This addresses the bug where exposed UDP ports always returned "0" instead of the actual mapped port.
//
// Background: When using ExposedPorts: []string{"8080/udp"}, the MappedPort() function
// would return "0/udp" instead of the actual host port like "55051/udp".
//
// Root cause: nat.ParsePortSpecs() creates PortBinding with empty HostPort (""),
// but Docker needs HostPort: "0" for automatic port allocation.
//
// Fix: In mergePortBindings(), convert empty HostPort to "0" for auto-allocation.
func TestUDPPortBinding(t *testing.T) {
	ctx := context.Background()

	t.Run("UDP port gets proper host port allocation", func(t *testing.T) {
		// Create container with UDP port exposed
		req := ContainerRequest{
			Image:        "alpine/socat:latest",
			ExposedPorts: []string{"8080/udp"},
			Cmd:          []string{"UDP-LISTEN:8080,fork,reuseaddr", "EXEC:'/bin/cat'"},
		}

		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, container.Terminate(ctx))
		}()

		// Test MappedPort function - this was the bug
		udpPort := "8080/udp"
		mappedPort, err := container.MappedPort(ctx, udpPort)
		require.NoError(t, err)

		// Before fix: mappedPort.Port() would return "0"
		// After fix: mappedPort.Port() returns actual port like "55051"
		assert.NotEqual(t, "0", mappedPort.Port(), "UDP port should not return '0'")
		assert.Equal(t, network.UDP, mappedPort.Proto(), "Protocol should be UDP")

		portNum := mappedPort.Num()
		assert.Positive(t, portNum, "Port number should be greater than 0")

		// Verify the port is actually accessible (basic connectivity test)
		hostIP, err := container.Host(ctx)
		require.NoError(t, err)

		address := net.JoinHostPort(hostIP, mappedPort.Port())
		conn, err := net.DialTimeout("udp", address, 2*time.Second)
		require.NoError(t, err, "Should be able to connect to UDP port")
		conn.Close()
	})

	t.Run("TCP port continues to work (regression test)", func(t *testing.T) {
		// Ensure our UDP fix doesn't break TCP ports
		req := ContainerRequest{
			Image:        "nginx:alpine",
			ExposedPorts: []string{"80/tcp"},
		}

		container, err := GenericContainer(ctx, GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, container.Terminate(ctx))
		}()

		tcpPort := "80/tcp"

		mappedPort, err := container.MappedPort(ctx, tcpPort)
		require.NoError(t, err)

		assert.NotEqual(t, "0", mappedPort.Port(), "TCP port should not return '0'")
		assert.Equal(t, network.TCP, mappedPort.Proto(), "Protocol should be TCP")

		portNum := mappedPort.Num()
		assert.Positive(t, portNum, "Port number should be greater than 0")
	})
}

// TestPortBindingInternalLogic tests the internal mergePortBindings function
// that was modified to fix the UDP port binding issue.
func TestPortBindingInternalLogic(t *testing.T) {
	t.Run("mergePortBindings fixes empty HostPort", func(t *testing.T) {
		// Test the core fix: empty HostPort should become "0"
		// This simulates what nat.ParsePortSpecs returns for "8080/udp"
		port := network.MustParsePort("8080/udp")
		exposedPortSet := network.PortSet{
			port: struct{}{},
		}
		configPortMap := network.PortMap{
			port: []network.PortBinding{{HostPort: ""}}, // Empty HostPort (the bug)
		}

		// Call the function our fix modified
		result := mergePortBindings(configPortMap, exposedPortSet)

		// Verify the fix worked
		require.Contains(t, result, port)
		bindings := result[port]
		require.Len(t, bindings, 1)

		// THE KEY ASSERTION: Empty HostPort should become "0"
		assert.Equal(t, "0", bindings[0].HostPort,
			"Empty HostPort should be converted to '0' for auto-allocation")
		assert.Zero(t, bindings[0].HostIP, "HostIP should remain empty for all interfaces")
	})

	t.Run("mergePortBindings preserves existing HostPort", func(t *testing.T) {
		// Ensure we don't modify already-set HostPort values
		port := network.MustParsePort("8080/udp")
		exposedPortSet := network.PortSet{
			port: struct{}{},
		}
		configPortMap := network.PortMap{
			port: []network.PortBinding{{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "9090"}},
		}

		result := mergePortBindings(configPortMap, exposedPortSet)

		bindings := result[port]
		require.Len(t, bindings, 1)

		// Should preserve existing values
		assert.Equal(t, "9090", bindings[0].HostPort, "Existing HostPort should be preserved")
		assert.Equal(t, "127.0.0.1", bindings[0].HostIP.String(), "Existing HostIP should be preserved")
	})

	t.Run("nat.ParsePortSpecs behavior documentation", func(t *testing.T) {
		// This test documents the behavior of nat.ParsePortSpecs that caused the bug
		exposedPorts := []string{"8080/udp", "9090/tcp"}
		exposedPortSet := network.PortSet{
			network.MustParsePort(exposedPorts[0]): struct{}{},
			network.MustParsePort(exposedPorts[1]): struct{}{},
		}
		configPortMap := network.PortMap{
			network.MustParsePort(exposedPorts[0]): []network.PortBinding{{HostPort: ""}},
			network.MustParsePort(exposedPorts[1]): []network.PortBinding{{HostPort: ""}},
		}

		// Call mergePortBindings which normalizes empty HostPort
		result := mergePortBindings(configPortMap, exposedPortSet)

		// Verify the port set
		assert.Contains(t, exposedPortSet, network.MustParsePort("8080/udp"))
		assert.Contains(t, exposedPortSet, network.MustParsePort("9090/tcp"))

		// Document the problematic behavior: nat.ParsePortSpecs creates empty HostPort
		udpBindings := result[network.MustParsePort("8080/udp")]
		require.Len(t, udpBindings, 1)
		assert.Equal(t, "0", udpBindings[0].HostPort,
			"Empty HostPort should be converted to '0' for auto-allocation")

		tcpBindings := result[network.MustParsePort("9090/tcp")]
		require.Len(t, tcpBindings, 1)
		assert.Equal(t, "0", tcpBindings[0].HostPort,
			"Empty HostPort should be converted to '0' for auto-allocation")
	})
}
