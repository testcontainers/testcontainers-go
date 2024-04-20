package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	createTables []dynamodb.CreateTableInput
}

type Option func(o *options)

func WithCreateTable(table dynamodb.CreateTableInput) Option {
	return func(o *options) {
		o.createTables = append(o.createTables, table)
	}
}

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) {
	// NOOP to satisfy interface.
}

func newOptions(opts ...testcontainers.ContainerCustomizer) options {
	o := new(options)

	for _, opt := range opts {
		if applyOpt, ok := opt.(Option); ok {
			applyOpt(o)
		}
	}

	return *o
}
