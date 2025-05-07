package toxiproxy

import (
	"fmt"
	"net"
	"strconv"
)

// proxy represents a single proxy configuration in the toxiproxy config file
type proxy struct {
	Name     string `json:"name"`
	Listen   string `json:"listen"`
	Upstream string `json:"upstream"`
	Enabled  bool   `json:"enabled"`

	listenIP     string
	upstreamIP   string
	listenPort   int
	upstreamPort int
}

// sanitize is a helper function that returns another proxy
// in which all the fields have been extracted from the
// string representation of the upstream and listen fields.
func (p *proxy) sanitize() error {
	listenIP, listenPortStr, err := net.SplitHostPort(p.Listen)
	if err != nil {
		return fmt.Errorf("split hostPort: %w", err)
	}
	p.listenIP = listenIP

	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		return fmt.Errorf("atoi: %w", err)
	}
	p.listenPort = listenPort

	upstreamIP, upstreamPortStr, err := net.SplitHostPort(p.Upstream)
	if err != nil {
		return fmt.Errorf("split hostPort: %w", err)
	}
	p.upstreamIP = upstreamIP

	upstreamPort, err := strconv.Atoi(upstreamPortStr)
	if err != nil {
		return fmt.Errorf("atoi: %w", err)
	}
	p.upstreamPort = upstreamPort

	return nil
}

// newProxy creates a new proxy configuration with default values
func newProxy(name string, upstream string) (*proxy, error) {
	_, upstreamPortStr, err := net.SplitHostPort(upstream)
	if err != nil {
		return nil, fmt.Errorf("split hostPort: %w", err)
	}

	_, err = strconv.Atoi(upstreamPortStr)
	if err != nil {
		return nil, fmt.Errorf("atoi: %w", err)
	}

	return &proxy{
		Name:     name,
		Upstream: upstream,
		Enabled:  true,
	}, nil
}
