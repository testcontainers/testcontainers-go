package redpanda

import (
	"fmt"
	"net"
	"strconv"

	"github.com/testcontainers/testcontainers-go"
)

// HTTPProxyAuthMethod defines the authentication method for HTTP Proxy.
type HTTPProxyAuthMethod string

const (
	HTTPProxyAuthMethodNone      HTTPProxyAuthMethod = "none"
	HTTPProxyAuthMethodHTTPBasic HTTPProxyAuthMethod = "http_basic"
	HTTPProxyAuthMethodOIDC      HTTPProxyAuthMethod = "oidc"
)

type options struct {
	// Superusers is a list of service account names.
	Superusers []string

	// KafkaEnableAuthorization is a flag to require authorization for Kafka connections.
	KafkaEnableAuthorization bool

	// KafkaAuthenticationMethod is either "none" for plaintext or "sasl"
	// for SASL (scram sha 256) authentication.
	KafkaAuthenticationMethod string

	// SchemaRegistryAuthenticationMethod is either "none" for no authentication
	// or "http_basic" for HTTP basic authentication.
	SchemaRegistryAuthenticationMethod string

	// HTTPProxyAuthenticationMethod is the authentication method for HTTP Proxy (pandaproxy).
	// Valid values are "none", "http_basic", or "oidc".
	HTTPProxyAuthenticationMethod HTTPProxyAuthMethod

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
		HTTPProxyAuthenticationMethod:      HTTPProxyAuthMethodNone,
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
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithNewServiceAccount includes a new user with username (key) and password (value)
// that shall be created, so that you can use these to authenticate against
// Redpanda (either for the Kafka API or Schema Registry HTTP access).
func WithNewServiceAccount(username, password string) Option {
	return func(o *options) error {
		o.ServiceAccounts[username] = password
		return nil
	}
}

// WithSuperusers defines the superusers added to the redpanda config.
// By default, there are no superusers.
func WithSuperusers(superusers ...string) Option {
	return func(o *options) error {
		o.Superusers = superusers
		return nil
	}
}

// WithEnableSASL enables SASL scram sha 256 authentication.
// By default, no authentication (plaintext) is used.
// When setting an authentication method, make sure to add users
// as well as authorize them using the WithSuperusers() option.
func WithEnableSASL() Option {
	return func(o *options) error {
		o.KafkaAuthenticationMethod = "sasl"
		return nil
	}
}

// WithEnableKafkaAuthorization enables authorization for connections on the Kafka API.
func WithEnableKafkaAuthorization() Option {
	return func(o *options) error {
		o.KafkaEnableAuthorization = true
		return nil
	}
}

// WithEnableWasmTransform enables wasm transform.
// Should not be used with RP versions before 23.3
func WithEnableWasmTransform() Option {
	return func(o *options) error {
		o.EnableWasmTransform = true
		return nil
	}
}

// WithEnableSchemaRegistryHTTPBasicAuth enables HTTP basic authentication for
// Schema Registry.
func WithEnableSchemaRegistryHTTPBasicAuth() Option {
	return func(o *options) error {
		o.SchemaRegistryAuthenticationMethod = "http_basic"
		return nil
	}
}

// WithHTTPProxyAuthMethod sets the authentication method for HTTP Proxy.
// If an invalid method is provided, it defaults to "none".
func WithHTTPProxyAuthMethod(method HTTPProxyAuthMethod) Option {
	switch method {
	case HTTPProxyAuthMethodNone, HTTPProxyAuthMethodHTTPBasic, HTTPProxyAuthMethodOIDC:
		return func(o *options) error {
			o.HTTPProxyAuthenticationMethod = method
			return nil
		}
	default:
		return func(o *options) error {
			// Invalid method, default to "none"
			o.HTTPProxyAuthenticationMethod = HTTPProxyAuthMethodNone
			return nil
		}
	}
}

// WithAutoCreateTopics enables topic auto creation.
func WithAutoCreateTopics() Option {
	return func(o *options) error {
		o.AutoCreateTopics = true
		return nil
	}
}

func WithTLS(cert, key []byte) Option {
	return func(o *options) error {
		o.EnableTLS = true
		o.cert = cert
		o.key = key
		return nil
	}
}

// WithListener adds a custom listener to the Redpanda containers. Listener
// will be aliases to all networks, so they can be accessed from within docker
// networks. At least one network must be attached to the container, if not an
// error will be thrown when starting the container.
func WithListener(lis string) Option {
	return func(o *options) error {
		host, port, err := net.SplitHostPort(lis)
		if err != nil {
			return fmt.Errorf("split host port: %w", err)
		}

		portInt, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("parse port: %w", err)
		}

		o.Listeners = append(o.Listeners, listener{
			Address:              host,
			Port:                 portInt,
			AuthenticationMethod: o.KafkaAuthenticationMethod,
		})
		return nil
	}
}

// WithBootstrapConfig adds an arbitrary config kvp to the Redpanda container.
// Per the name, this config will be interpolated into the generated bootstrap
// config file, which is particularly useful for configs requiring a restart
// when otherwise applied to a running Redpanda instance.
func WithBootstrapConfig(cfg string, val any) Option {
	return func(o *options) error {
		o.ExtraBootstrapConfig[cfg] = val
		return nil
	}
}

// WithAdminAPIAuthentication enables Admin API Authentication.
// It sets `admin_api_require_auth` configuration to true and configures a bootstrap user account.
// See https://docs.redpanda.com/current/deploy/deployment-option/self-hosted/manual/production/production-deployment/#bootstrap-a-user-account
func WithAdminAPIAuthentication() Option {
	return func(o *options) error {
		o.enableAdminAPIAuthentication = true
		return nil
	}
}
