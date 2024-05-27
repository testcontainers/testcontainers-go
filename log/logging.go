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

type standardLogger struct {
	*log.Logger
}

// Printf implements Logging.
func (s standardLogger) Printf(format string, v ...interface{}) {
	s.Logger.Printf(format, v...)
}

// StandardLogger returns a default Logging implementation using the standard log package.
func StandardLogger() Logging {
	return standardLogger{Logger: log.Default()}
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
