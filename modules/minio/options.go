package minio

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"MINIO_ROOT_USER":     defaultUser,
			"MINIO_ROOT_PASSWORD": defaultPassword,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Minio container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a user and its password.
// It will create the specified user. It must not be empty or undefined.
func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["MINIO_ROOT_USER"] = username

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is required for you to use the Minio image. It must not be empty or undefined.
// This environment variable sets the root user password for Minio.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["MINIO_ROOT_PASSWORD"] = password

		return nil
	}
}
