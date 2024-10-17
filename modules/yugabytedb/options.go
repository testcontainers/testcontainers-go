package yugabytedb

import (
	"github.com/testcontainers/testcontainers-go"
)

// WithDatabaseName sets the initial database name for the yugabyteDB container.
func WithDatabaseName(dbName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseNameEnv] = dbName
		return nil
	}
}

// WithDatabaseUser sets the initial database user for the yugabyteDB container.
func WithDatabaseUser(dbUser string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseUserEnv] = dbUser
		return nil
	}
}

// WithDatabasePassword sets the initial database password for the yugabyteDB container.
func WithDatabasePassword(dbPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabasePasswordEnv] = dbPassword
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
