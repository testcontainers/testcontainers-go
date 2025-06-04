package yugabytedb

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	envs map[string]string
}

func defaultOptions() options {
	return options{
		envs: map[string]string{
			ycqlKeyspaceEnv:         ycqlKeyspace,
			ycqlUserNameEnv:         ycqlUserName,
			ycqlPasswordEnv:         ycqlPassword,
			ysqlDatabaseNameEnv:     ysqlDatabaseName,
			ysqlDatabaseUserEnv:     ysqlDatabaseUser,
			ysqlDatabasePasswordEnv: ysqlDatabasePassword,
		},
	}
}

// Compiler check to ensure that Option implements the testcontainers.ContainerCustomizer interface.
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the Redpanda container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithDatabaseName sets the initial database name for the yugabyteDB container.
func WithDatabaseName(dbName string) Option {
	return func(o *options) error {
		o.envs[ysqlDatabaseNameEnv] = dbName
		return nil
	}
}

// WithDatabaseUser sets the initial database user for the yugabyteDB container.
func WithDatabaseUser(dbUser string) Option {
	return func(o *options) error {
		o.envs[ysqlDatabaseUserEnv] = dbUser
		return nil
	}
}

// WithDatabasePassword sets the initial database password for the yugabyteDB container.
func WithDatabasePassword(dbPassword string) Option {
	return func(o *options) error {
		o.envs[ysqlDatabasePasswordEnv] = dbPassword
		return nil
	}
}

// WithKeyspace sets the initial keyspace for the yugabyteDB container.
func WithKeyspace(keyspace string) Option {
	return func(o *options) error {
		o.envs[ycqlKeyspaceEnv] = keyspace
		return nil
	}
}

// WithUser sets the initial user for the yugabyteDB container.
func WithUser(user string) Option {
	return func(o *options) error {
		o.envs[ycqlUserNameEnv] = user
		return nil
	}
}

// WithPassword sets the initial password for the yugabyteDB container.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.envs[ycqlPasswordEnv] = password
		return nil
	}
}
