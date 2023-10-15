package elasticsearch

import (
	"github.com/testcontainers/testcontainers-go"
)

// Options is a struct for specifying options for the Elasticsearch container.
// It could be used to build an HTTP client for the Elasticsearch container, as it will
// hold information on how to connect to the container.
type Options struct {
	Address  string
	CACert   []byte
	Password string
	Username string
}

func defaultOptions() *Options {
	return &Options{
		CACert:   nil,
		Username: defaultUsername,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the Elasticsearch container.
type Option func(*Options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithPassword sets the password for the Elasticsearch container.
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}
