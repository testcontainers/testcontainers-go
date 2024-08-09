package core

import (
	"context"
	"fmt"
	"net"
	"sync"
)

type IPFamily string

const (
	IPv4 IPFamily = "IPv4"
	IPv6 IPFamily = "IPv6"
)

var (
	hostIPs     []HostIP
	hostIPsOnce sync.Once
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
// The function is protected by a sync.Once to avoid unnecessary calculations.
func GetDockerHostIPs() []HostIP {
	hostIPsOnce.Do(func() {
		dockerHost := ExtractDockerHost(context.Background())
		hostIPs = getDockerHostIPs(dockerHost)
	})

	return hostIPs
}

// getDockerHostIPs returns the IP addresses of the Docker host.
// The function is helpful for testing purposes,
// as it's not protected by the sync.Once.
func getDockerHostIPs(host string) []HostIP {
	hip := newHostIP(host)

	ips, err := net.LookupIP(hip.Address)
	if err != nil {
		return []HostIP{hip}
	}

	hips := []HostIP{}
	for _, ip := range ips {
		hips = append(hips, newHostIP(ip.String()))
	}

	return hips
}
