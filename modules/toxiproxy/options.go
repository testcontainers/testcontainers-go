package toxiproxy

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	proxies []*proxy
}

func defaultOptions() options {
	return options{
		proxies: []*proxy{},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithProxy creates a new proxy configuration for the given name and upstream.
// If the upstream is not a valid IP address and port, it returns an error.
// If this option is used in combination with the [WithConfigFile] option, the proxy defined in this option
// is added to the existing proxies.
func WithProxy(name string, upstream string) Option {
	return func(o *options) error {
		proxy, err := newProxy(name, upstream)
		if err != nil {
			return fmt.Errorf("newProxy: %w", err)
		}

		o.proxies = append(o.proxies, proxy)
		return nil
	}
}

// WithConfigFile sets the config file for the Toxiproxy container, copying
// the file to the "/tmp/tc-toxiproxy.json" path. It also appends the "-host=0.0.0.0"
// and "-config=/tmp/tc-toxiproxy.json" flags to the command line.
// The config file is a JSON file that contains the configuration for the Toxiproxy container,
// and it is not validated by the Toxiproxy container.
// If this option is used in combination with the [WithProxy] option, the proxies defined in this option
// are added to the existing proxies.
func WithConfigFile(r io.Reader) Option {
	return func(o *options) error {
		// unmarshal the config file
		var config []proxy
		err := json.NewDecoder(r).Decode(&config)
		if err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}

		for _, proxy := range config {
			proxy, err := newProxy(proxy.Name, proxy.Upstream)
			if err != nil {
				return fmt.Errorf("newProxy: %w", err)
			}

			o.proxies = append(o.proxies, proxy)
		}

		return nil
	}
}
