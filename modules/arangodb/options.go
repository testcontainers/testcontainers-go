package arangodb

import "github.com/testcontainers/testcontainers-go"

// WithRootPassword sets the password for the ArangoDB root user
func WithRootPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ARANGO_ROOT_PASSWORD"] = password
		return nil
	}
}
