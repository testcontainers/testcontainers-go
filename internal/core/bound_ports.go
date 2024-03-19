package core

import (
	"fmt"

	"github.com/docker/go-connections/nat"
)

// ResolveHostPortBinding resolves the host port binding for the host IPs
func ResolveHostPortBinding(hostIPs []HostIP, portBindings []nat.PortBinding) (int, error) {
	for _, hp := range hostIPs {
		family := hp.Family

		for _, portBinding := range portBindings {
			hostIP := newHostIP(portBinding.HostIP)
			if hostIP.Family == family {
				return nat.Port(portBinding.HostPort).Int(), nil
			}
		}
	}

	return 0, fmt.Errorf("no host port found for host IPs %v", hostIPs)
}
