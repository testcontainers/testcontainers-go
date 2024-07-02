package network

import "github.com/docker/docker/api/types/network"

type Request struct {
	Driver     string
	Internal   bool
	EnableIPv6 *bool
	Name       string
	Labels     map[string]string
	Attachable bool
	IPAM       *network.IPAM
}
