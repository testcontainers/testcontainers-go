package core

import (
	"fmt"

	"github.com/docker/go-connections/nat"
)

type BoundPorts map[nat.Port]nat.Port

// BoundPortsFromBindings returns a map of container ports to host ports.
// They are resolved from the port bindings in the inspect response,
// using the host IP addresses of the Docker host.
// This will resolve the issue of the host port being bound to multiple IP addresses
// in the IPv4 and IPv6 case.
func BoundPortsFromBindings(portMap nat.PortMap) (BoundPorts, error) {
	hostIPs := GetDockerHostIPs()

	boundPorts := make(BoundPorts)

	for containerPort, bindings := range portMap {
		if len(bindings) == 0 {
			continue
		}

		hostPort, err := resolveHostPortBinding(hostIPs, bindings)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve host port binding for port %s: %w", containerPort, err)
		}

		boundPorts[containerPort] = hostPort
	}

	return boundPorts, nil
}

// resolveHostPortBinding resolves the host port binding for the host IPs.
// It will return the host port for the first matching IP family (IPv4 or IPv6).
func resolveHostPortBinding(hostIPs []HostIP, portBindings []nat.PortBinding) (nat.Port, error) {
	for _, hp := range hostIPs {
		family := hp.Family

		for _, portBinding := range portBindings {
			hostIP := newHostIP(portBinding.HostIP)
			if hostIP.Family == family {
				return nat.Port(portBinding.HostPort), nil
			}
		}
	}

	return "", fmt.Errorf("no host port found for host IPs %v", hostIPs)
}
