package trinogateway

import (
	"errors"
	"io"

	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	dataStore DataStoreProvider
	config    io.Reader
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

// WithDataStore sets the data store of the gateway, see NewManagedPostgresProvider and
// NewExternalProvider. It is required.
func WithDataStore(provider DataStoreProvider) Option {
	return func(o *options) error {
		if provider == nil {
			return errors.New("data store provider must not be nil")
		}

		if err := provider.validate(); err != nil {
			return err
		}

		o.dataStore = provider
		return nil
	}
}

// WithConfig replaces the default config file of the gateway. The data store settings are in the
// DATASTORE_JDBC_URL, DATASTORE_USER, DATASTORE_PASSWORD and DATASTORE_DRIVER env vars.
func WithConfig(r io.Reader) Option {
	return func(o *options) error {
		if r == nil {
			return errors.New("config reader must not be nil")
		}

		o.config = r
		return nil
	}
}
