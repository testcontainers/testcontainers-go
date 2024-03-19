package core

import (
	"fmt"
	"net"
)

type IPFamily string

const (
	IPv4 IPFamily = "IPv4"
	IPv6 IPFamily = "IPv6"
)

type HostIP struct {
	Address string
	Family  IPFamily
}

func (h HostIP) String() string {
	return fmt.Sprintf("%s (%s)", h.Address, h.Family)
}

func newHostIP(host string) HostIP {
	var hip HostIP

	ip := net.ParseIP(host)
	if ip == nil {
		host = "127.0.0.1"
		ip = net.ParseIP(host)
	}

	hip.Address = host

	if ip.To4() != nil {
		hip.Family = IPv4
	} else if ip.To16() != nil {
		hip.Family = IPv6
	}

	return hip
}

// GetDockerHostIPs returns the IP addresses of the Docker host.
func GetDockerHostIPs(host string) []HostIP {
	hip := newHostIP(host)

	ips, err := net.LookupIP(hip.Address)
	if err != nil {
		return []HostIP{hip}
	}

	var hostIPs []HostIP
	for _, ip := range ips {
		hostIPs = append(hostIPs, newHostIP(ip.String()))
	}

	return hostIPs
}
