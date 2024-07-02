package postgres

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BasicWaitStrategies is a simple but reliable way to wait for postgres to start.
// It returns a two-step wait strategy:
//
//   - It will wait for the container to log `database system is ready to accept connections` twice, because it will restart itself after the first startup.
//   - It will then wait for docker to actually serve the port on localhost.
//     For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
//     Without this, the tests will be flaky on those OSes!
func BasicWaitStrategies() testcontainers.CustomizeRequestOption {
	// waitStrategy {
	return testcontainers.WithWaitStrategy(
		// First, we wait for the container to log readiness twice.
		// This is because it will restart itself after the first startup.
		wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		// Then, we wait for docker to actually serve the port on localhost.
		// For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
		// Without this, the tests will be flaky on those OSes!
		wait.ForListeningPort("5432/tcp"),
	)
	// }
}
