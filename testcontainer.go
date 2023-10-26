package testcontainers

import "github.com/testcontainers/testcontainers-go/internal/testcontainerssession"

// SessionID returns a unique session ID for the current test session. Because each Go package
// will be run in a separate process, we need a way to identify the current test session.
// By test session, we mean:
//   - a single "go test" invocation (including flags)
//   - a single "go test ./..." invocation (including flags)
//   - the execution of a single test or a set of tests using the IDE
//
// As a consequence, with the sole goal of aggregating test execution across multiple
// packages, this variable will contain the value of the parent process ID (pid) of the current process
// and its creation date, to use it to generate a unique session ID. We are using the parent pid because
// the current process will be a child process of:
//   - the process that is running the tests, e.g.: "go test";
//   - the process that is running the application in development mode, e.g. "go run main.go -tags dev";
//   - the process that is running the tests in the IDE, e.g.: "go test ./...".
//
// Finally, we will hash the combination of the "testcontainers-go:" string with the parent pid
// and the creation date of that parent process to generate a unique session ID.
//
// This sessionID will be used to:
//   - identify the test session, aggregating the test execution of multiple packages in the same test session.
//   - tag the containers created by testcontainers-go, adding a label to the container with the session ID.
func SessionID() string {
	return testcontainerssession.SessionID()
}
