package dex

import (
	"errors"
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
	clients                 []Client
	users                   []User
	connectors              []connector
	issuer                  string
	skipApprovalScreen      bool
	storage                 Storage
	logLevel                slog.Level
	logger                  *slog.Logger
	enablePasswordDB        bool
	enableClientCredentials bool
}

func defaultOptions() options {
	return options{
		skipApprovalScreen: true,
		storage:            StorageSQLite,
		logLevel:           slog.LevelInfo,
		enablePasswordDB:   true,
	}
}

// Option is a functional option for the Dex module. Options return an error
// so user-supplied values can be validated at Run time rather than failing
// silently in the rendered YAML.
//
// NOTE: Options must be passed directly to dex.Run. They satisfy the
// testcontainers.ContainerCustomizer interface only so the Run signature
// can accept them alongside generic tc-go customizers (e.g. network.With*)
// — Option.Customize is a no-op, so an Option forwarded through any
// wrapper that dispatches via Customize (instead of type-asserting to
// Option) is silently dropped.
type Option func(*options) error

// Compiler check: Option implements testcontainers.ContainerCustomizer.
var _ testcontainers.ContainerCustomizer = Option(nil)

// Customize is a no-op; real state mutation happens inside Run. See the
// Option type-level doc for why this is a no-op.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	return nil
}

// WithClient registers a static client in Dex's YAML config. Unlike
// gRPC-added clients, these may declare custom grant types.
func WithClient(c Client) Option {
	return func(o *options) error {
		if c.id == "" {
			return errors.New("dex: WithClient requires a Client constructed via NewClient")
		}
		o.clients = append(o.clients, c)
		return nil
	}
}

// WithUser registers a static password entry. The password DB connector is
// enabled by default, so no extra option is needed to consume the entry.
func WithUser(u User) Option {
	return func(o *options) error {
		if u.email == "" {
			return errors.New("dex: WithUser requires a User constructed via NewUser")
		}
		o.users = append(o.users, u)
		return nil
	}
}

// WithConnector enables a Dex connector by type. For ConnectorPassword this
// is a no-op — the password DB is enabled by default and the template
// handles it separately; id and name are ignored and blank-field
// validation is skipped in that case. For other connectors
// (e.g. ConnectorMock) the entry is added to the rendered YAML, and blank
// id or name returns an error.
func WithConnector(t ConnectorType, id, name string) Option {
	return func(o *options) error {
		if t == ConnectorPassword {
			return nil
		}
		if id == "" {
			return errors.New("dex: connector id must not be blank")
		}
		if name == "" {
			return errors.New("dex: connector name must not be blank")
		}
		o.connectors = append(o.connectors, connector{Type: t, ID: id, Name: name})
		return nil
	}
}

// WithIssuer overrides the default host:mappedPort-derived issuer. When set,
// Run uses the fast-path (direct YAML bind-mount). Callers are responsible
// for ensuring the URL is reachable from every client (tests and sibling
// containers).
func WithIssuer(url string) Option {
	return func(o *options) error {
		if url == "" {
			return errors.New("dex: issuer URL must not be blank")
		}
		o.issuer = url
		return nil
	}
}

// WithSkipApprovalScreen toggles Dex's oauth2.skipApprovalScreen. Default: true.
func WithSkipApprovalScreen(skip bool) Option {
	return func(o *options) error {
		o.skipApprovalScreen = skip
		return nil
	}
}

// WithStorage sets Dex's storage backend. Default: StorageSQLite.
func WithStorage(s Storage) Option {
	return func(o *options) error {
		if s == "" {
			return errors.New("dex: storage must not be blank")
		}
		o.storage = s
		return nil
	}
}

// WithDisablePasswordDB disables Dex's built-in password connector. The
// caller must then configure at least one other connector via WithConnector,
// otherwise Run returns ErrNoAuthSource.
func WithDisablePasswordDB() Option {
	return func(o *options) error {
		o.enablePasswordDB = false
		return nil
	}
}

// WithLogger routes Dex container logs through the supplied slog.Logger.
// When unset, Dex container logs are discarded. Calling WithLogger(nil)
// is a no-op; to discard logs again after setting a logger, drop the
// option rather than passing nil.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) error {
		if logger == nil {
			return nil
		}
		o.logger = logger
		return nil
	}
}

// WithLogLevel sets Dex's own --log-level flag. Accepts a standard library
// slog.Level; values are mapped to Dex's level vocabulary (debug, info,
// warn, error). Default: slog.LevelInfo.
func WithLogLevel(level slog.Level) Option {
	return func(o *options) error {
		o.logLevel = level
		return nil
	}
}

// WithEnableClientCredentials enables Dex's OAuth2 client_credentials grant
// via the DEX_CLIENT_CREDENTIAL_GRANT_ENABLED_BY_DEFAULT=true environment
// variable.
//
// Requires Dex ≥ v2.46.0 or the dexidp/dex:master image tag. Earlier
// releases silently ignore the flag and token exchanges fail with
// unsupported_grant_type. This module does not validate the image tag —
// the caller must pin a compatible image.
func WithEnableClientCredentials() Option {
	return func(o *options) error {
		o.enableClientCredentials = true
		return nil
	}
}
