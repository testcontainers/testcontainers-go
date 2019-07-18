package testcontainers

import "context"

// Network allows getting info about a single network instance
type Network interface {
	Remove(context.Context) error // removes the network
}

// NetworkRequest represents the parameters used to get a network
type NetworkRequest struct {
	Driver         string
	CheckDuplicate bool
	Internal       bool
	EnableIPv6     bool
	Name           string
	Attachable     bool
}
