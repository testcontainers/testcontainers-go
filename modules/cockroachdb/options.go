package cockroachdb

import "github.com/testcontainers/testcontainers-go"

type options struct {
	Database   string
	User       string
	Password   string
	StoreSize  string
	TLS        *TLSConfig
	Statements []string
}

func defaultOptions() options {
	return options{
		User:      defaultUser,
		Password:  defaultPassword,
		Database:  defaultDatabase,
		StoreSize: defaultStoreSize,
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

// ClusterDefaults are the settings recommended by Cockroach Labs for testing clusters.
// See https://www.cockroachlabs.com/docs/stable/local-testing for more information.
var ClusterDefaults []string = []string{
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
func WithStatements(statements ...string) Option {
	return func(o *options) {
		o.Statements = statements
	}
}
