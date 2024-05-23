package log

import (
	"log"
	"testing"
)

// Validate our types implement the required interfaces.
var (
	_ Logging = (*log.Logger)(nil)
)

// Logging defines the Logger interface
type Logging interface {
	Printf(format string, v ...interface{})
}

type testLogger struct {
	testing.TB
}

// Printf implements Logging.
func (t testLogger) Printf(format string, v ...interface{}) {
	t.Helper()
	t.Logf(format, v...)
}

// TestLogger returns a Logging implementation for testing.TB
// This way logs from testcontainers are part of the test output of a test suite or test case.
func NewTestLogger(tb testing.TB) Logging {
	tb.Helper()
	return testLogger{TB: tb}
}
