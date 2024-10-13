package meilisearch

import (
	"github.com/testcontainers/testcontainers-go"
	"path/filepath"
)

// Options is a struct for specifying options for the Meilisearch container.
type Options struct {
	DumpDataFileDir  string
	DumpDataFileName string
}

func defaultOptions() *Options {
	return &Options{}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (*Option)(nil)

// Option is an option for the Meilisearch container.
type Option func(*Options)

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithDumpImport sets the data dump file path for the Meilisearch container.
func WithDumpImport(dumpFilePath string) Option {
	return func(o *Options) {
		o.DumpDataFileDir, o.DumpDataFileName = filepath.Dir(dumpFilePath), filepath.Base(dumpFilePath)
	}
}
