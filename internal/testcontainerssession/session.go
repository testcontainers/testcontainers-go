package testcontainerssession

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// SessionID returns a unique session ID for the current test session. Because each Go package
// will be run in a separate process, we need a way to identify the current test session.
// By test session, we mean:
//   - a single "go test" invocation (including flags)
//   - a single "go test ./..." invocation (including flags)
//   - the execution of a single test or a set of tests using the IDE
//
// As a consequence, with the sole goal of aggregating test execution across multiple
// packages, this function will use the parent process ID (pid) of the current process
// and use it to generate a unique session ID. We are using the parent pid because
// the current process will be a child process of:
//   - the process that is running the tests, e.g.: "go test";
//   - the process that is running the application in development mode, e.g. "go run main.go -tags dev";
//   - the process that is running the tests in the IDE, e.g.: "go test ./...".
//
// Finally, we will hash the combination of the "testcontainers-go:" string and the parent pid
// to generate a unique session ID.
//
// This SessionID will be used to:
//   - identify the test session, aggregating the test execution of multiple packages in the same test session.
//   - tag the containers created by testcontainers-go, adding a label to the container with the session ID.
var SessionID string

// RunID returns a unique run ID for the current test process. Because each Go package will be run in a separate process,
// we need a way to identify the current test process.
var RunID string

const sessionIDPlaceholder = "testcontainers-go:%d"

func init() {
	parentPid := os.Getppid()

	hasher := sha256.New()
	_, err := hasher.Write([]byte(fmt.Sprintf(sessionIDPlaceholder, parentPid)))
	if err != nil {
		SessionID = uuid.New().String()
		return
	}

	SessionID = fmt.Sprintf("%x", hasher.Sum(nil))
	RunID = uuid.New().String()
}
