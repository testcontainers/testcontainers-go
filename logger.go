package testcontainers

import (
	"log"
	"os"
	"testing"
)

// Logger is the default log instance
var Logger Logging = log.New(os.Stderr, "", log.LstdFlags)

// Logging defines the Logger interface
type Logging interface {
	Printf(format string, v ...interface{})
}

func TestLogger(tb testing.TB) Logging {
	return testLogger{TB: tb}
}

type testLogger struct {
	testing.TB
}

func (t testLogger) Printf(format string, v ...interface{}) {
	t.Logf(format, v...)
}
