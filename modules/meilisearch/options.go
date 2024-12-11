package meilisearch

import (
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
)

// Options is a struct for specifying options for the Meilisearch container.
type Options struct {
	DumpDataFilePath string
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
// dumpFilePath either relative to where you call meilisearch run or absolute path
func WithDumpImport(dumpFilePath string) Option {
	return func(o *Options) {
		o.DumpDataFilePath, o.DumpDataFileName = dumpFilePath, filepath.Base(dumpFilePath)
	}
}

// WithMasterKey sets the master key for the Meilisearch container
// it satisfies the testcontainers.ContainerCustomizer interface
func WithMasterKey(masterKey string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["MEILI_MASTER_KEY"] = masterKey
		return nil
	}
}
