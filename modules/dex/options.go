package dex

import (
	"log/slog"

	"github.com/testcontainers/testcontainers-go"
)

// connector is an internal record of a WithConnector call.
type connector struct {
	Type ConnectorType
	ID   string
	Name string
}

// options is the module-internal accumulator for Run().
type options struct {
	clients                  []Client
	users                    []User
	connectors               []connector
	issuer                   string // empty means derive from host:mapped at post-start
	skipApprovalScreen       bool
	storage                  string
	logLevel                 string
	logger                   *slog.Logger
	enablePasswordDB         bool // default true; flipped off only if user sets ConnectorMock-only
	enableClientCredentials  bool
}

func defaultOptions() options {
	return options{
		skipApprovalScreen: true,
		storage:            "sqlite3",
		logLevel:           "info",
		enablePasswordDB:   true,
	}
}

// Compiler check: Option implements testcontainers.ContainerCustomizer.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is a functional option for the Dex module.
type Option func(*options)

// Customize is a no-op; Option satisfies the testcontainers.ContainerCustomizer
// interface so callers can pass Options through any API that accepts
// tc-go customizers. Real state mutation happens in Run().
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	return nil
}

// WithClient registers a static client in Dex's YAML config. Unlike
// gRPC-added clients, these may declare custom grant types.
func WithClient(c Client) Option {
	return func(o *options) {
		o.clients = append(o.clients, c)
	}
}

// WithUser registers a static password entry. The password DB connector is
// enabled by default, so no extra option is needed to consume the entry.
func WithUser(u User) Option {
	return func(o *options) {
		o.users = append(o.users, u)
	}
}

// WithConnector enables a Dex connector by type. For ConnectorPassword, this
// is a no-op (the password DB is enabled by default whenever WithUser is
// passed or no other connector is configured). For ConnectorMock, the
// mockCallback connector is added.
func WithConnector(t ConnectorType, id, name string) Option {
	return func(o *options) {
		// ConnectorPassword is handled via enablePasswordDB in the YAML template;
		// appending it to o.connectors would emit a spurious `type: password`
		// entry that Dex does not recognize.
		if t == ConnectorPassword {
			return
		}
		o.connectors = append(o.connectors, connector{Type: t, ID: id, Name: name})
	}
}

// WithIssuer overrides the default host:mappedPort-derived issuer. When set,
// Run uses the fast-path (direct YAML bind-mount). Callers are responsible
// for ensuring the URL is reachable from every client (tests and sibling
// containers).
func WithIssuer(url string) Option {
	return func(o *options) {
		o.issuer = url
	}
}

// WithSkipApprovalScreen toggles Dex's oauth2.skipApprovalScreen. Default: true.
func WithSkipApprovalScreen(skip bool) Option {
	return func(o *options) {
		o.skipApprovalScreen = skip
	}
}

// WithStorage sets Dex's storage backend. Default: "sqlite3". "memory" is
// also supported by Dex.
func WithStorage(kind string) Option {
	return func(o *options) {
		o.storage = kind
	}
}

// WithLogger routes Dex container logs through the supplied slog.Logger.
// When nil or not called, Dex container logs are discarded.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithLogLevel sets Dex's own --log-level flag. Valid: "debug", "info",
// "warn", "error". Default: "info".
func WithLogLevel(level string) Option {
	return func(o *options) {
		o.logLevel = level
	}
}

// WithEnableClientCredentials enables Dex's OAuth2 client_credentials grant
// via the DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true environment
// variable. Requires Dex ≥ v2.46.0 (not yet released at time of writing)
// or the dexidp/dex:master image tag; earlier releases do not recognize
// the flag. The module logs a warning when an older image tag is
// detected. See the module README for image compatibility.
func WithEnableClientCredentials() Option {
	return func(o *options) {
		o.enableClientCredentials = true
	}
}
