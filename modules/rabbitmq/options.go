package rabbitmq

import "github.com/testcontainers/testcontainers-go"

type SSLVerificationMode string

const (
	SSLVerificationModeNone SSLVerificationMode = "verify_none"
	SSLVerificationModePeer SSLVerificationMode = "verify_peer"
)

type options struct {
	AdminUsername string
	AdminPassword string
	SSLSettings   *SSLSettings
}

func defaultOptions() options {
	return options{
		AdminUsername: defaultUser,
		AdminPassword: defaultPassword,
	}
}

type SSLSettings struct {
	// Path to the CA certificate file
	CACertFile string
	// Path to the client certificate file
	CertFile string
	// Path to the key file
	KeyFile string
	// Verification mode
	VerificationMode SSLVerificationMode
	// Fail if no certificate is provided
	FailIfNoCert bool
	// Depth of certificate chain verification
	VerificationDepth int
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the RabbitMQ container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithAdminPassword sets the password for the default admin user
func WithAdminPassword(password string) Option {
	return func(o *options) {
		o.AdminPassword = password
	}
}

// WithAdminUsername sets the default admin username
func WithAdminUsername(username string) Option {
	return func(o *options) {
		o.AdminUsername = username
	}
}

// WithSSL enables SSL on the RabbitMQ container, configuring the Erlang config file with the provided settings.
func WithSSL(settings SSLSettings) Option {
	return func(o *options) {
		o.SSLSettings = &settings
	}
}
