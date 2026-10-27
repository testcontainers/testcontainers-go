package trinogateway

import (
	"errors"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	postgresImage   string
	postgresOptions []testcontainers.ContainerCustomizer
	dataStore       *dataStore
	config          io.Reader
}

// dataStore holds the settings of the database used by Trino Gateway as data store.
type dataStore struct {
	jdbcURL  string
	user     string
	password string
	driver   string
}

func defaultOptions() options {
	return options{
		postgresImage: defaultPostgresImage,
	}
}

// Satisfy the testcontainers.ContainerCustomizer interface
var _ testcontainers.ContainerCustomizer = Option(nil)

// Option is an option for the Trino Gateway container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithPostgres sets the image and options for the PostgreSQL container used as data store.
// By default, the image is "postgres:17-alpine".
// This option is ignored when an external data store is configured with WithDataStore.
func WithPostgres(img string, opts ...testcontainers.ContainerCustomizer) Option {
	return func(o *options) error {
		o.postgresImage = img
		o.postgresOptions = opts
		return nil
	}
}

// WithDataStore configures Trino Gateway to use an external database as data store,
// instead of starting a PostgreSQL container. The JDBC URL must be reachable from the
// Trino Gateway container, and the driver is the JDBC driver class, e.g. "org.postgresql.Driver",
// "com.mysql.cj.jdbc.Driver" or "oracle.jdbc.OracleDriver".
func WithDataStore(jdbcURL, user, password, driver string) Option {
	return func(o *options) error {
		if jdbcURL == "" {
			return errors.New("data store JDBC URL must not be empty")
		}

		o.dataStore = &dataStore{
			jdbcURL:  jdbcURL,
			user:     user,
			password: password,
			driver:   driver,
		}
		return nil
	}
}

// WithConfig replaces the default configuration file of Trino Gateway with the content
// of the reader, copying it to the container file at "/etc/trino-gateway/config.yaml".
// The configuration can reference the data store settings using the environment variables
// set by the module: "${ENV:DATASTORE_JDBC_URL}", "${ENV:DATASTORE_USER}",
// "${ENV:DATASTORE_PASSWORD}" and "${ENV:DATASTORE_DRIVER}".
func WithConfig(r io.Reader) Option {
	return func(o *options) error {
		if r == nil {
			return errors.New("config reader must not be nil")
		}

		o.config = r
		return nil
	}
}
