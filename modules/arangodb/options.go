package arangodb

import (
	"net/http"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// WithRootPassword sets the password for the ArangoDB root user
func WithRootPassword(password string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Env["ARANGO_ROOT_PASSWORD"] = password
		return nil
	}
}

// withWaitStrategy sets the wait strategy for the ArangoDB container
// once we know the credentials.
func withWaitStrategy() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.WaitingFor = wait.ForAll(
			req.WaitingFor,
			wait.ForHTTP("/_admin/status").
				WithPort(defaultPort).
				WithBasicAuth(DefaultUser, req.Env["ARANGO_ROOT_PASSWORD"]).
				WithHeaders(map[string]string{
					"Accept": "application/json",
				}).
				WithStatusCodeMatcher(func(status int) bool {
					return status == http.StatusOK
				}),
		)

		return nil
	}
}
