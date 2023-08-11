package testcontainers

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/docker/docker/client"
	"github.com/testcontainers/testcontainers-go/internal/testcontainersdocker"
)

// Logger is the default log instance
var Logger Logging = log.New(os.Stderr, "", log.LstdFlags)

// Logging defines the Logger interface
type Logging interface {
	Printf(format string, v ...interface{})
}

// LogDockerServerInfo logs the docker server info using the provided logger and Docker client
func LogDockerServerInfo(ctx context.Context, client client.APIClient, logger Logging) {
	infoMessage := `%v - Connected to docker: 
  Server Version: %v
  API Version: %v
  Operating System: %v
  Total Memory: %v MB
  Resolved Docker Host: %s
  Resolved Docker Socket Path: %s
`

	info, err := client.Info(ctx)
	if err != nil {
		logger.Printf("failed getting information about docker server: %s", err)
		return
	}

	logger.Printf(infoMessage, packagePath,
		info.ServerVersion, client.ClientVersion(),
		info.OperatingSystem, info.MemTotal/1024/1024,
		testcontainersdocker.ExtractDockerHost(ctx),
		testcontainersdocker.ExtractDockerSocket(ctx),
	)
}

// TestLogger returns a Logging implementation for testing.TB
// This way logs from testcontainers are part of the test output of a test suite or test case
func TestLogger(tb testing.TB) Logging {
	tb.Helper()
	return testLogger{TB: tb}
}

// WithLogger is a generic option that implements GenericProviderOption, DockerProviderOption
// It replaces the global Logging implementation with a user defined one e.g. to aggregate logs from testcontainers
// with the logs of specific test case
func WithLogger(logger Logging) LoggerOption {
	return LoggerOption{
		logger: logger,
	}
}

type LoggerOption struct {
	logger Logging
}

func (o LoggerOption) ApplyGenericTo(opts *GenericProviderOptions) {
	opts.Logger = o.logger
}

func (o LoggerOption) ApplyDockerTo(opts *DockerProviderOptions) {
	opts.Logger = o.logger
}

type testLogger struct {
	testing.TB
}

func (t testLogger) Printf(format string, v ...interface{}) {
	t.Helper()
	t.Logf(format, v...)
}
