package testcontainers

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/require"
)

// SkipIfProviderIsNotHealthy is a utility function capable of skipping tests
// if the provider is not healthy, or running at all.
// This is a function designed to be used in your test, when Docker is not mandatory for CI/CD.
// In this way tests that depend on Testcontainers won't run if the provider is provisioned correctly.
func SkipIfProviderIsNotHealthy(t *testing.T) {
	ctx := context.Background()
	provider, err := ProviderDocker.GetProvider()
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
	err = provider.Health(ctx)
	if err != nil {
		t.Skipf("Docker is not running. TestContainers can't perform is work without it: %s", err)
	}
}

// SkipIfDockerDesktop is a utility function capable of skipping tests
// if tests are run using Docker Desktop.
func SkipIfDockerDesktop(t *testing.T, ctx context.Context) {
	cli, err := NewDockerClientWithOpts(ctx)
	if err != nil {
		t.Fatalf("failed to create docker client: %s", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		t.Fatalf("failed to get docker info: %s", err)
	}

	if info.OperatingSystem == "Docker Desktop" {
		t.Skip("Skipping test that requires host network access when running in Docker Desktop")
	}
}

// exampleLogConsumer {

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct{}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l Log) {
	fmt.Print(string(l.Content))
}

// }

// CleanupContainer is a helper function that schedules the container
// to be stopped / terminated when the test ends.
//
// This should be called as a defer directly after (before any error check)
// of [GenericContainer](...) or a modules Run(...) in a test to ensure the
// container is stopped when the function ends.
//
// before any error check. If container is nil, its a no-op.
func CleanupContainer(tb testing.TB, container Container, options ...TerminateOption) {
	tb.Helper()

	tb.Cleanup(func() {
		noErrorOrNotFound(tb, TerminateContainer(container, options...))
	})
}

// CleanupNetwork is a helper function that schedules the network to be
// removed when the test ends.
// This should be the first call after NewNetwork(...) in a test before
// any error check. If network is nil, its a no-op.
func CleanupNetwork(tb testing.TB, network Network) {
	tb.Helper()

	tb.Cleanup(func() {
		noErrorOrNotFound(tb, network.Remove(context.Background()))
	})
}

// noErrorOrNotFound is a helper function that checks if the error is nil or a not found error.
func noErrorOrNotFound(tb testing.TB, err error) {
	tb.Helper()

	if isNilOrNotFound(err) {
		return
	}

	require.NoError(tb, err)
}

// causer is an interface that allows to get the cause of an error.
type causer interface {
	Cause() error
}

// wrapErr is an interface that allows to unwrap an error.
type wrapErr interface {
	Unwrap() error
}

// unwrapErrs is an interface that allows to unwrap multiple errors.
type unwrapErrs interface {
	Unwrap() []error
}

// isNilOrNotFound reports whether all errors in err's tree are either nil or implement [errdefs.ErrNotFound].
func isNilOrNotFound(err error) bool {
	if err == nil {
		return true
	}

	switch x := err.(type) { //nolint:errorlint // We need to check for interfaces.
	case errdefs.ErrNotFound:
		return true
	case causer:
		return isNilOrNotFound(x.Cause())
	case wrapErr:
		return isNilOrNotFound(x.Unwrap())
	case unwrapErrs:
		for _, e := range x.Unwrap() {
			if !isNilOrNotFound(e) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
