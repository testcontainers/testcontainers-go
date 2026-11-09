package postgres

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// BasicWaitStrategies is a simple but reliable way to wait for postgres to start.
// It returns a three-step wait strategy:
//
//   - It will wait for the container to log `database system is ready to accept connections` twice, because it will restart itself after the first startup.
//   - It will then probe the live server state with `pg_isready` until postgres accepts connections.
//     Container logs survive restarts, so on a reused container the log message of a previous run
//     could otherwise report readiness before the current process accepts connections
//     (https://github.com/testcontainers/testcontainers-go/issues/3671). The probe prefers the
//     unix socket, falls back to TCP on loopback honoring PGPORT, needs no valid credentials,
//     and is skipped on images that do not ship pg_isready, preserving the previous behavior.
//   - It will then wait for docker to actually serve the port on localhost.
//     For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
//     Without this, the tests will be flaky on those OSes!
func BasicWaitStrategies() testcontainers.CustomizeRequestOption {
	// waitStrategy {
	return testcontainers.WithAdditionalWaitStrategy(
		// First, we wait for the container to log readiness twice.
		// This is because it will restart itself after the first startup.
		wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		// Then, we probe the live server state until it accepts connections, because
		// the log message of a previous run of a reused container cannot prove that
		// the current process is ready. Skipped on images without pg_isready.
		wait.ForExec([]string{"sh", "-c", `command -v pg_isready >/dev/null 2>&1 || exit 0; pg_isready || pg_isready -h 127.0.0.1 -p "${PGPORT:-5432}"`}),
		// Then, we wait for docker to actually serve the port on localhost.
		// For non-linux OSes like Mac and Windows, Docker or Rancher Desktop will have to start a separate proxy.
		// Without this, the tests will be flaky on those OSes!
		wait.ForListeningPort("5432/tcp"),
	)
	// }
}
