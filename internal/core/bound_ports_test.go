package core

import (
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
)

func TestResolveHostPortBinding(t *testing.T) {
	type testCase struct {
		name         string
		expectedPort nat.Port
		hostIPs      []HostIP
		bindings     []nat.PortBinding
		expectedErr  error
	}

	testCases := []testCase{
		{
			name: "should return IPv6-mapped host port when preferred",
			hostIPs: []HostIP{
				{Family: IPv6, Address: "::1"},
				{Family: IPv4, Address: "127.0.0.1"},
			},
			bindings: []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: "50000"},
				{HostIP: "::", HostPort: "50001"},
			},
			expectedPort: nat.Port("50001"),
		},
		{
			name: "should return IPv4-mapped host port when preferred",
			hostIPs: []HostIP{
				{Family: IPv4, Address: "127.0.0.1"},
				{Family: IPv6, Address: "::1"},
			},
			bindings: []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: "50000"},
				{HostIP: "::", HostPort: "50001"},
			},
			expectedPort: nat.Port("50000"),
		},
		{
			name: "should return mapped host port when dual stack IP",
			hostIPs: []HostIP{
				{Family: IPv4, Address: "127.0.0.1"},
				{Family: IPv6, Address: "::1"},
			},
			bindings: []nat.PortBinding{
				{HostIP: "", HostPort: "50000"},
			},
			expectedPort: nat.Port("50000"),
		},
		{
			name: "should throw when no host port available for host IP family",
			hostIPs: []HostIP{
				{Family: IPv6, Address: "::1"},
			},
			bindings: []nat.PortBinding{
				{HostIP: "0.0.0.0", HostPort: "50000"},
			},
			expectedPort: nat.Port(""), // that's the zero value returned by ResolveHostPortBinding
			expectedErr:  fmt.Errorf("no host port found for host IPs [%s (IPv6)]", "::1"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolvedPort, err := resolveHostPortBinding(tc.hostIPs, tc.bindings)

			switch {
			case err == nil && tc.expectedErr == nil:
				break
			case err == nil && tc.expectedErr != nil:
				t.Errorf("did not receive expected error: %s", tc.expectedErr.Error())
				return
			case err != nil && tc.expectedErr == nil:
				t.Errorf("unexpected error: %v", err)
				return
			case err.Error() != tc.expectedErr.Error():
				t.Errorf("errors mismatch: %s != %s", err.Error(), tc.expectedErr.Error())
				return
			}

			if resolvedPort != tc.expectedPort {
				t.Errorf("resolved port mismatch: got %s, expected %s", resolvedPort, tc.expectedPort)
			}
		})
	}
}
