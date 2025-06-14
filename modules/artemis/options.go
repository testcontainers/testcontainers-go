package artemis

import "github.com/testcontainers/testcontainers-go"

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{
			"ARTEMIS_USER":     "artemis",
			"ARTEMIS_PASSWORD": "artemis",
		},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Artemis container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithCredentials sets the administrator credentials. The default is artemis:artemis.
func WithCredentials(user, password string) Option {
	return func(o *options) error {
		o.env["ARTEMIS_USER"] = user
		o.env["ARTEMIS_PASSWORD"] = password

		return nil
	}
}

// WithAnonymousLogin enables anonymous logins.
func WithAnonymousLogin() Option {
	return func(o *options) error {
		o.env["ANONYMOUS_LOGIN"] = "true"

		return nil
	}
}

// Additional arguments sent to the `artemis create` command.
// The default is `--http-host 0.0.0.0 --relax-jolokia`.
// Setting this value will override the default.
// See the documentation on `artemis create` for available options.
func WithExtraArgs(args string) Option {
	return func(o *options) error {
		o.env["EXTRA_ARGS"] = args

		return nil
	}
}
