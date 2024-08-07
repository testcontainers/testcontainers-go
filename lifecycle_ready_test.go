package testcontainers

import (
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestPortMappingCheck(t *testing.T) {
	makePortMap := func(ports ...string) nat.PortMap {
		out := make(nat.PortMap)
		for _, port := range ports {
			// We don't care about the actual binding in this test
			out[nat.Port(port)] = nil
		}
		return out
	}

	tests := map[string]struct {
		exposedAndMappedPorts nat.PortMap
		exposedPorts          []string
		expectError           bool
	}{
		"no-protocol": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024"},
		},
		"protocol": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024/tcp"},
		},
		"protocol-target-port": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024:1024/tcp"},
		},
		"target-port": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"1024:1024"},
		},
		"multiple-ports": {
			exposedAndMappedPorts: makePortMap("1024/tcp", "1025/tcp", "1026/tcp"),
			exposedPorts:          []string{"1024", "25:1025/tcp", "1026:1026"},
		},
		"only-ipv4": {
			exposedAndMappedPorts: makePortMap("1024/tcp"),
			exposedPorts:          []string{"0.0.0.0::1024/tcp"},
		},
		"no-mapped-ports": {
			exposedAndMappedPorts: makePortMap(),
			exposedPorts:          []string{"1024"},
			expectError:           true,
		},
		"wrong-mapped-port": {
			exposedAndMappedPorts: makePortMap("1023/tcp"),
			exposedPorts:          []string{"1024"},
			expectError:           true,
		},
		"subset-mapped-ports": {
			exposedAndMappedPorts: makePortMap("1024/tcp", "1025/tcp"),
			exposedPorts:          []string{"1024", "1025", "1026"},
			expectError:           true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkPortsMapped(tt.exposedAndMappedPorts, tt.exposedPorts)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
