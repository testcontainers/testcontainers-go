package toxiproxy

import (
	"fmt"
	"net"
	"strconv"
)

// proxy represents a single proxy configuration in the toxiproxy config file
type proxy struct {
	Name         string `json:"name"`
	Listen       string `json:"listen"`
	listenIP     string
	listenPort   int
	Upstream     string `json:"upstream"`
	upstreamIP   string
	upstreamPort int
	Enabled      bool `json:"enabled"`
}

// sanitize is a helper function that returns another proxy
// in which all the fields have been extracted from the
// string representation of the upstream and listen fields.
func (p proxy) sanitize() (proxy, error) {
	var pp proxy

	listenIP, listenPortStr, err := net.SplitHostPort(p.Listen)
	if err != nil {
		return pp, fmt.Errorf("split hostPort: %w", err)
	}

	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		return pp, fmt.Errorf("atoi: %w", err)
	}

	upstreamIP, upstreamPortStr, err := net.SplitHostPort(p.Upstream)
	if err != nil {
		return pp, fmt.Errorf("split hostPort: %w", err)
	}

	upstreamPort, err := strconv.Atoi(upstreamPortStr)
	if err != nil {
		return pp, fmt.Errorf("atoi: %w", err)
	}

	pp = proxy{
		Name:         p.Name,
		Listen:       p.Listen,
		listenIP:     listenIP,
		listenPort:   listenPort,
		Upstream:     p.Upstream,
		upstreamIP:   upstreamIP,
		upstreamPort: upstreamPort,
		Enabled:      p.Enabled,
	}

	return pp, nil
}

// newProxy creates a new proxy configuration with default values
func newProxy(name string, upstream string) (proxy, error) {
	_, upstreamPortStr, err := net.SplitHostPort(upstream)
	if err != nil {
		return proxy{}, fmt.Errorf("split hostPort: %w", err)
	}

	_, err = strconv.Atoi(upstreamPortStr)
	if err != nil {
		return proxy{}, fmt.Errorf("atoi: %w", err)
	}

	return proxy{
		Name:     name,
		Upstream: upstream,
		Enabled:  true,
	}, nil
}
