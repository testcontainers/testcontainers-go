// Package fakesnow — see package documentation in fakesnow.go.
package fakesnow

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

// options holds the connection details [Container.ConnectionString] builds a DSN from.
// fakesnow accepts any credentials, so these only affect what the driver sends.
type options struct {
	account  string
	user     string
	password string
	database string
	schema   string
}

// defaultOptions returns an options struct with the default configuration.
func defaultOptions() options {
	return options{
		account:  "fakesnow",
		user:     "fakesnow",
		password: "fakesnow",
		database: "db1",
		schema:   "schema1",
	}
}

// Satisfy the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = Option(nil)

// Option is an option for the fakesnow container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithAccount sets the account in the DSN. fakesnow accepts any account, so this
// only changes what the driver sends. It cannot be empty: gosnowflake then falls
// back to deriving the account from the host, which fails for a container address.
func WithAccount(account string) Option {
	return func(o *options) error {
		if account == "" {
			return fmt.Errorf("account cannot be empty")
		}
		o.account = account

		return nil
	}
}

// WithUsername sets the user in the DSN. fakesnow accepts any user.
func WithUsername(user string) Option {
	return func(o *options) error {
		o.user = user

		return nil
	}
}

// WithPassword sets the password in the DSN. fakesnow accepts any password.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.password = password

		return nil
	}
}

// WithDatabase sets the database in the DSN. The database is not created for you,
// run CREATE DATABASE IF NOT EXISTS yourself before using it.
func WithDatabase(database string) Option {
	return func(o *options) error {
		o.database = database

		return nil
	}
}

// WithSchema sets the schema in the DSN. The schema is not created for you,
// run CREATE SCHEMA IF NOT EXISTS yourself before using it.
func WithSchema(schema string) Option {
	return func(o *options) error {
		o.schema = schema

		return nil
	}
}
