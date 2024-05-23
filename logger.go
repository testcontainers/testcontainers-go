package testcontainers

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/docker/docker/client"

	tclog "github.com/testcontainers/testcontainers-go/log"
)

// Logger is the default log instance
var Logger tclog.Logging = log.New(os.Stderr, "", log.LstdFlags)

// Validate our types implement the required interfaces.
var (
	_ Logging               = (*log.Logger)(nil) // Deprecated: use tclog.Logging instead
	_ ContainerCustomizer   = LoggerOption{}
	_ GenericProviderOption = LoggerOption{}
	_ DockerProviderOption  = LoggerOption{}
)

// Deprecated: use tclog.Logging instead
// Logging defines the Logger interface
type Logging = tclog.Logging

// Deprecated: this function will be removed in a future release
// LogDockerServerInfo logs the docker server info using the provided logger and Docker client
func LogDockerServerInfo(ctx context.Context, client client.APIClient, logger Logging) {
	// NOOP
}

// Deprecated: use log.NewTestLogger instead
// TestLogger returns a Logging implementation for testing.TB
// This way logs from testcontainers are part of the test output of a test suite or test case.
func TestLogger(tb testing.TB) tclog.Logging {
	tb.Helper()
	return testLogger{TB: tb}
}

// WithLogger returns a generic option that sets the logger to be used.
//
// Consider calling this before other "With functions" as these may generate logs.
//
// This can be given a TestLogger to collect the logs from testcontainers into a
// test case.
func WithLogger(logger Logging) LoggerOption {
	return LoggerOption{
		logger: logger,
	}
}

// LoggerOption is a generic option that sets the logger to be used.
//
// It can be used to set the logger for providers and containers.
type LoggerOption struct {
	logger Logging
}

// ApplyGenericTo implements GenericProviderOption.
func (o LoggerOption) ApplyGenericTo(opts *GenericProviderOptions) {
	opts.Logger = o.logger
}

// ApplyDockerTo implements DockerProviderOption.
func (o LoggerOption) ApplyDockerTo(opts *DockerProviderOptions) {
	opts.Logger = o.logger
}

// Customize implements ContainerCustomizer.
func (o LoggerOption) Customize(req *GenericContainerRequest) error {
	req.Logger = o.logger
	return nil
}

// Deprecated: use log.testLogger instead
type testLogger struct {
	testing.TB
}

// Deprecated: use log.testLogger instead
// Printf implements Logging.
func (t testLogger) Printf(format string, v ...interface{}) {
	t.Helper()
	t.Logf(format, v...)
}
