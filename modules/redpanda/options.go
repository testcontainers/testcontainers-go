package redpanda

import (
	"net"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// Superusers is a list of service account names.
	Superusers []string

	// KafkaEnableAuthorization is a flag to require authorization for Kafka connections.
	KafkaEnableAuthorization bool

	// KafkaAuthenticationMethod is either "none" for plaintext or "sasl"
	// for SASL (scram) authentication.
	KafkaAuthenticationMethod string

	// SchemaRegistryAuthenticationMethod is either "none" for no authentication
	// or "http_basic" for HTTP basic authentication.
	SchemaRegistryAuthenticationMethod string

	// EnableWasmTransform is a flag to enable wasm transform.
	EnableWasmTransform bool

	// ServiceAccounts is a map of username (key) to password (value) of users
	// that shall be created, so that you can use these to authenticate against
	// Redpanda (either for the Kafka API or Schema Registry HTTP access).
	// You must use SCRAM-SHA-256 as algorithm when authenticating on the
	// Kafka API.
	ServiceAccounts map[string]string

	// AutoCreateTopics is a flag to allow topic auto creation.
	AutoCreateTopics bool

	// EnableTLS is a flag to enable TLS.
	EnableTLS bool

	cert, key []byte

	// Listeners is a list of custom listeners that can be provided to access the
	// containers form within docker networks
	Listeners []listener

	// ExtraBootstrapConfig is a map of configs to be interpolated into the
	// container's bootstrap.yml
	ExtraBootstrapConfig map[string]any

	// enableAdminAPIAuthentication enables Admin API authentication
	enableAdminAPIAuthentication bool
}

func defaultOptions() options {
	return options{
		Superusers:                         []string{},
		KafkaEnableAuthorization:           false,
		KafkaAuthenticationMethod:          "none",
		SchemaRegistryAuthenticationMethod: "none",
		ServiceAccounts:                    make(map[string]string, 0),
		AutoCreateTopics:                   false,
		EnableTLS:                          false,
		Listeners:                          make([]listener, 0),
		ExtraBootstrapConfig:               make(map[string]any, 0),
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

func WithNewServiceAccount(username, password string) Option {
	return func(o *options) {
		o.ServiceAccounts[username] = password
	}
}

// WithSuperusers defines the superusers added to the redpanda config.
// By default, there are no superusers.
func WithSuperusers(superusers ...string) Option {
	return func(o *options) {
		o.Superusers = superusers
	}
}

// WithEnableSASL enables SASL scram sha authentication.
// By default, no authentication (plaintext) is used.
// When setting an authentication method, make sure to add users
// as well as authorize them using the WithSuperusers() option.
func WithEnableSASL() Option {
	return func(o *options) {
		o.KafkaAuthenticationMethod = "sasl"
	}
}

// WithEnableKafkaAuthorization enables authorization for connections on the Kafka API.
func WithEnableKafkaAuthorization() Option {
	return func(o *options) {
		o.KafkaEnableAuthorization = true
	}
}

// WithEnableWasmTransform enables wasm transform.
// Should not be used with RP versions before 23.3
func WithEnableWasmTransform() Option {
	return func(o *options) {
		o.EnableWasmTransform = true
	}
}

// WithEnableSchemaRegistryHTTPBasicAuth enables HTTP basic authentication for
// Schema Registry.
func WithEnableSchemaRegistryHTTPBasicAuth() Option {
	return func(o *options) {
		o.SchemaRegistryAuthenticationMethod = "http_basic"
	}
}

// WithAutoCreateTopics enables topic auto creation.
func WithAutoCreateTopics() Option {
	return func(o *options) {
		o.AutoCreateTopics = true
	}
}

func WithTLS(cert, key []byte) Option {
	return func(o *options) {
		o.EnableTLS = true
		o.cert = cert
		o.key = key
	}
}

// WithListener adds a custom listener to the Redpanda containers. Listener
// will be aliases to all networks, so they can be accessed from within docker
// networks. At least one network must be attached to the container, if not an
// error will be thrown when starting the container.
func WithListener(lis string) Option {
	host, port, err := net.SplitHostPort(lis)
	if err != nil {
		return func(_ *options) {}
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return func(_ *options) {}
	}

	return func(o *options) {
		o.Listeners = append(o.Listeners, listener{
			Address:              host,
			Port:                 portInt,
			AuthenticationMethod: o.KafkaAuthenticationMethod,
		})
	}
}

// WithBootstrapConfig adds an arbitrary config kvp to the Redpanda container.
// Per the name, this config will be interpolated into the generated bootstrap
// config file, which is particularly useful for configs requiring a restart
// when otherwise applied to a running Redpanda instance.
func WithBootstrapConfig(cfg string, val any) Option {
	return func(o *options) {
		o.ExtraBootstrapConfig[cfg] = val
	}
}

// WithAdminAPIAuthentication enables Admin API Authentication.
// It sets `admin_api_require_auth` configuration to true and configures a bootstrap user account.
// See https://docs.redpanda.com/current/deploy/deployment-option/self-hosted/manual/production/production-deployment/#bootstrap-a-user-account
func WithAdminAPIAuthentication() Option {
	return func(o *options) {
		o.enableAdminAPIAuthentication = true
	}
}
