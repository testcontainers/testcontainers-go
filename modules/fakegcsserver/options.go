// Package fakegcsserver — see package documentation in fakegcsserver.go.
package fakegcsserver

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// options holds the configuration for the FakeGCSServer container.
type options struct {
	// Scheme is the HTTP scheme to use. Valid values are "http" and "https".
	Scheme string
}

// defaultOptions returns an options struct with the default configuration.
func defaultOptions() options {
	return options{
		Scheme: "http",
	}
}

// Satisfy the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the FakeGCSServer container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithScheme sets the scheme used by the fake-gcs-server.
// Valid values are "http" (default) and "https".
//
// Note: "-scheme both" is not supported because it requires a second port (default 8000)
// for HTTP in addition to the HTTPS port (4443) and this module only exposes 4443/tcp.
func WithScheme(scheme string) Option {
	return func(o *options) error {
		switch scheme {
		case "http", "https":
			o.Scheme = scheme
		default:
			return fmt.Errorf("invalid scheme %q: must be one of http, https", scheme)
		}

		return nil
	}
}
