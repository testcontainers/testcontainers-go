package yugabytedb

import (
	"github.com/testcontainers/testcontainers-go"
)

// WithYSQLDatabaseName sets the initial database name for the yugabyteDB container.
func WithYSQLDatabaseName(ysqlDBName string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseNameEnv] = ysqlDBName
		return nil
	}
}

// WithYSQLDatabaseUser sets the initial database user for the yugabyteDB container.
func WithYSQLDatabaseUser(ysqlDBUser string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabaseUserEnv] = ysqlDBUser
		return nil
	}
}

// WithYSQLDatabasePassword sets the initial database password for the yugabyteDB container.
func WithYSQLDatabasePassword(ysqlDBPassword string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ysqlDatabasePasswordEnv] = ysqlDBPassword
		return nil
	}
}

// WithYCQLKeyspace sets the initial keyspace for the yugabyteDB container.
func WithYCQLKeyspace(keyspace string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlKeyspaceEnv] = keyspace
		return nil
	}
}

// WithYCQLUser sets the initial user for the yugabyteDB container.
func WithYCQLUser(user string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlUserNameEnv] = user
		return nil
	}
}

// WithYCQLPassword sets the initial password for the yugabyteDB container.
func WithYCQLPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env[ycqlPasswordEnv] = password
		return nil
	}
}
