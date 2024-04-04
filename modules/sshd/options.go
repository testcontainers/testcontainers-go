package sshd

import (
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	RootPassword string

	Options []string
}

func defaultOptions() options {
	return options{
		RootPassword: uuid.NewString(),

		Options: []string{},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the SSHD container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithRootPassword sets the root password
func WithRootPassword(password string) Option {
	return func(o *options) {
		o.RootPassword = password
	}
}

// WithRootPassword sets the options
func WithOptions(opts []string) Option {
	return func(o *options) {
		o.Options = opts
	}
}
