package logging

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/client"
)

// Logger is the default log instance
var Logger Logging = &noopLogger{}

func init() {
	// Enable default logger in the testing with a verbose flag.
	if testing.Testing() {
		// Parse manually because testing.Verbose() panics unless flag.Parse() has done.
		for _, arg := range os.Args {
			if strings.EqualFold(arg, "-test.v=true") || strings.EqualFold(arg, "-v") {
				Logger = log.New(os.Stderr, "", log.LstdFlags)
			}
		}
	}
}

// Validate our types implement the required interfaces.
var (
	_ Logging = (*log.Logger)(nil)
)

// Logging defines the Logger interface
type Logging interface {
	Printf(format string, v ...any)
	Print(v ...any)
}

type noopLogger struct{}

// Printf implements Logging.
func (n noopLogger) Printf(format string, v ...any) {
	// NOOP
}

// Print implements Logging.
func (n noopLogger) Print(v ...any) {
	// NOOP
}

// Deprecated: this function will be removed in a future release
// LogDockerServerInfo logs the docker server info using the provided logger and Docker client
func LogDockerServerInfo(_ context.Context, _ client.APIClient, _ Logging) {
	// NOOP
}

// TestLogger returns a Logging implementation for testing.TB
// This way logs from testcontainers are part of the test output of a test suite or test case.
func TestLogger(tb testing.TB) Logging {
	tb.Helper()
	return testLogger{TB: tb}
}

type testLogger struct {
	testing.TB
}

// Printf implements Logging.
func (t testLogger) Printf(format string, v ...any) {
	t.Helper()
	t.Logf(format, v...)
}

// Print implements Logging.
func (t testLogger) Print(v ...any) {
	t.Helper()
	t.Log(v...)
}
