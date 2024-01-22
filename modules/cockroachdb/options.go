package cockroachdb

import "github.com/testcontainers/testcontainers-go"

// Options is a struct for specifying options for the CockroachDB container.
type Options struct {
	Database  string
	StoreSize string
}

func defaultOptions() Options {
	return Options{
		Database:  defaultDatabase,
		StoreSize: defaultStoreSize,
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the CockroachDB container.
type Option func(*Options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

// WithDatabase sets the name of the database to use.
func WithDatabase(database string) Option {
	return func(o *Options) {
		o.Database = database
	}
}

// WithStoreSize sets the amount of available in-memory storage.
// See https://www.cockroachlabs.com/docs/stable/cockroach-start#store
func WithStoreSize(size string) Option {
	return func(o *Options) {
		o.StoreSize = size
	}
}
