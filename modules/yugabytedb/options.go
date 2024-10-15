package yugabytedb

import (
	"github.com/testcontainers/testcontainers-go"
)

// WithDatabaseName sets the initial database name for the yugabyteDB container.
func WithDatabaseName(ysqlDBName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseNameEnv] = ysqlDBName
		return nil
	}
}

// WithDatabaseUser sets the initial database user for the yugabyteDB container.
func WithDatabaseUser(ysqlDBUser string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseUserEnv] = ysqlDBUser
		return nil
	}
}

// WithDatabasePassword sets the initial database password for the yugabyteDB container.
func WithDatabasePassword(ysqlDBPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabasePasswordEnv] = ysqlDBPassword
		return nil
	}
}

// WithKeyspace sets the initial keyspace for the yugabyteDB container.
func WithKeyspace(keyspace string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlKeyspaceEnv] = keyspace
		return nil
	}
}

// WithUser sets the initial user for the yugabyteDB container.
func WithUser(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlUserNameEnv] = user
		return nil
	}
}

// WithPassword sets the initial password for the yugabyteDB container.
func WithPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlPasswordEnv] = password
		return nil
	}
}
