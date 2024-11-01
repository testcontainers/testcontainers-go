package cockroachdb

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	Database   string
	User       string
	Password   string
	StoreSize  string
	TLS        *TLSConfig
	Statements []string
}

// containerConnConfig returns the [pgx.ConnConfig] for the given container and options.
func (opts options) containerConnConfig(ctx context.Context, container testcontainers.Container) (*pgx.ConnConfig, error) {
	port, err := container.MappedPort(ctx, defaultSQLPort)
	if err != nil {
		return nil, fmt.Errorf("mapped port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("host: %w", err)
	}

	return opts.connConfig(host, port)
}

// containerConnString returns the connection string for the given container and options.
func (opts options) containerConnString(ctx context.Context, container testcontainers.Container) (string, error) {
	cfg, err := opts.containerConnConfig(ctx, container)
	if err != nil {
		return "", fmt.Errorf("container connection config: %w", err)
	}

	return stdlib.RegisterConnConfig(cfg), nil
}

// connString returns a connection string for the given host, port and options.
func (opts options) connString(host string, port nat.Port) (string, error) {
	cfg, err := opts.connConfig(host, port)
	if err != nil {
		return "", fmt.Errorf("connection config: %w", err)
	}

	return stdlib.RegisterConnConfig(cfg), nil
}

// connConfig returns a [pgx.ConnConfig] for the given host, port and options.
func (opts options) connConfig(host string, port nat.Port) (*pgx.ConnConfig, error) {
	user := url.User(opts.User)
	if opts.Password != "" {
		user = url.UserPassword(opts.User, opts.Password)
	}

	sslMode := "disable"
	if opts.TLS != nil {
		sslMode = "require" // We can't use "verify-full" as it might be a self signed cert.
	}
	params := url.Values{
		"sslmode": []string{sslMode},
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     user,
		Host:     net.JoinHostPort(host, port.Port()),
		Path:     opts.Database,
		RawQuery: params.Encode(),
	}

	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if opts.TLS != nil {
		tlsCfg, err := opts.TLS.tlsConfig()
		if err != nil {
			return nil, fmt.Errorf("tls config: %w", err)
		}

		cfg.TLSConfig = tlsCfg
	}

	return cfg, nil
}

func defaultOptions() options {
	return options{
		User:       defaultUser,
		Password:   defaultPassword,
		Database:   defaultDatabase,
		StoreSize:  defaultStoreSize,
		Statements: DefaultStatements,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the CockroachDB container.
type Option func(*options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithDatabase sets the name of the database to use.
func WithDatabase(database string) Option {
	return func(o *options) {
		o.Database = database
	}
}

// WithUser creates & sets the user to connect as.
func WithUser(user string) Option {
	return func(o *options) {
		o.User = user
	}
}

// WithPassword sets the password when using password authentication.
func WithPassword(password string) Option {
	return func(o *options) {
		o.Password = password
	}
}

// WithStoreSize sets the amount of available in-memory storage.
// See https://www.cockroachlabs.com/docs/stable/cockroach-start#store
func WithStoreSize(size string) Option {
	return func(o *options) {
		o.StoreSize = size
	}
}

// WithTLS enables TLS on the CockroachDB container.
// Cert and key must be PEM-encoded.
func WithTLS(cfg *TLSConfig) Option {
	return func(o *options) {
		o.TLS = cfg
	}
}

// DefaultStatements are the settings recommended by Cockroach Labs for testing clusters.
// Note that to use these defaults the user needs to have MODIFYCLUSTERSETTING privilege.
// See https://www.cockroachlabs.com/docs/stable/local-testing for more information.
var DefaultStatements = []string{
	"SET CLUSTER SETTING kv.range_merge.queue_interval = '50ms'",
	"SET CLUSTER SETTING jobs.registry.interval.gc = '30s'",
	"SET CLUSTER SETTING jobs.registry.interval.cancel = '180s'",
	"SET CLUSTER SETTING jobs.retention_time = '15s'",
	"SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false",
	"SET CLUSTER SETTING kv.range_split.by_load_merge_delay = '5s'",
	`ALTER RANGE default CONFIGURE ZONE USING "gc.ttlseconds" = 600`,
	`ALTER DATABASE system CONFIGURE ZONE USING "gc.ttlseconds" = 600`,
}

// WithStatements sets the statements to run on the CockroachDB cluster once the container is ready.
// By default, the container will run the statements in [DefaultStatements] as recommended by
// Cockroach Labs however that is not always possible due to the user not having the required privileges.
func WithStatements(statements ...string) Option {
	return func(o *options) {
		o.Statements = statements
	}
}
