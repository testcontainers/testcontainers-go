package postgres

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	// SQLDriverName is the name of the SQL driver to use.
	SQLDriverName string
	Snapshot      string
}

func defaultOptions() options {
	return options{
		SQLDriverName: "postgres",
		Snapshot:      defaultSnapshotName,
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

// WithSQLDriver sets the SQL driver to use for the container.
// It is passed to sql.Open() to connect to the database when making or restoring snapshots.
// This can be set if your app imports a different postgres driver, f.ex. "pgx"
func WithSQLDriver(driver string) Option {
	return func(o *options) {
		o.SQLDriverName = driver
	}
}
