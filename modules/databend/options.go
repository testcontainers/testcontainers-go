package databend

import "github.com/testcontainers/testcontainers-go"

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"QUERY_DEFAULT_USER":     defaultUser,
			"QUERY_DEFAULT_PASSWORD": defaultPassword,
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Databend container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithUsername sets the username for the Databend container.
// WithUsername is [Run] option that configures the default query user by setting
// the `QUERY_DEFAULT_USER` container environment variable.
func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["QUERY_DEFAULT_USER"] = username
		return nil
	}
}

// WithPassword sets the password for the Databend container.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["QUERY_DEFAULT_PASSWORD"] = password
		return nil
	}
}
