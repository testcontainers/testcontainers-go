package mongodb

import (
	"github.com/testcontainers/testcontainers-go"
)

type options struct {
	env map[string]string
}

func defaultOptions() options {
	return options{
		env: map[string]string{},
	}
}

// Satisfy the testcontainers.CustomizeRequestOption interface
var _ testcontainers.ContainerCustomizer = (Option)(nil)

// Option is an option for the MongoDB container.
type Option func(*options) error

// Customize is a NOOP. It's defined to satisfy the testcontainers.ContainerCustomizer interface.
func (o Option) Customize(*testcontainers.GenericContainerRequest) error {
	// NOOP to satisfy interface.
	return nil
}

// WithUsername sets the initial username to be created when the container starts
// It is used in conjunction with WithPassword to set a username and its password.
// It will create the specified user with superuser power.
func WithUsername(username string) Option {
	return func(o *options) error {
		o.env["MONGO_INITDB_ROOT_USERNAME"] = username

		return nil
	}
}

// WithPassword sets the initial password of the user to be created when the container starts
// It is used in conjunction with WithUsername to set a username and its password.
// It will set the superuser password for MongoDB.
func WithPassword(password string) Option {
	return func(o *options) error {
		o.env["MONGO_INITDB_ROOT_PASSWORD"] = password

		return nil
	}
}

// WithReplicaSet sets the replica set name for Single node MongoDB replica set.
func WithReplicaSet(replSetName string) Option {
	return func(o *options) error {
		o.env[replicaSetOptEnvKey] = replSetName

		return nil
	}
}
