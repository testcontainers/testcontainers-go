package testcontainers

import (
	"context"
	"regexp"
	"testing"

	"github.com/docker/docker/errdefs"
	"github.com/stretchr/testify/require"
)

// errAlreadyInProgress is a regular expression that matches the error for a container
// removal that is already in progress.
var errAlreadyInProgress = regexp.MustCompile(`removal of container .* is already in progress`)

// CleanupContainer is a helper function that schedules the container
// to be stopped / terminated when the test ends.
//
// This should be called as a defer directly after (before any error check)
// of [GenericContainer](...) or a modules Run(...) in a test to ensure the
// container is stopped when the function ends.
//
// before any error check. If container is nil, its a no-op.
func CleanupContainer(tb testing.TB, ctr StartedContainer, options ...TerminateOption) {
	tb.Helper()

	tb.Cleanup(func() {
		noErrorOrIgnored(tb, TerminateContainer(ctr, options...))
	})
}

// CleanupNetwork is a helper function that schedules the network to be
// removed when the test ends.
// This should be the first call after NewNetwork(...) in a test before
// any error check. If network is nil, its a no-op.
func CleanupNetwork(tb testing.TB, network Network) {
	tb.Helper()

	tb.Cleanup(func() {
		noErrorOrIgnored(tb, network.Remove(context.Background()))
	})
}

// noErrorOrIgnored is a helper function that checks if the error is nil or an error
// we can ignore.
func noErrorOrIgnored(tb testing.TB, err error) {
	tb.Helper()

	if isCleanupSafe(err) {
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

// isCleanupSafe reports whether all errors in err's tree are one of the
// following, so can safely be ignored:
//   - nil
//   - not found
//   - already in progress
func isCleanupSafe(err error) bool {
	if err == nil {
		return true
	}

	switch x := err.(type) { //nolint:errorlint // We need to check for interfaces.
	case errdefs.ErrNotFound:
		return true
	case errdefs.ErrConflict:
		// Terminating a container that is already terminating.
		if errAlreadyInProgress.MatchString(err.Error()) {
			return true
		}
		return false
	case causer:
		return isCleanupSafe(x.Cause())
	case wrapErr:
		return isCleanupSafe(x.Unwrap())
	case unwrapErrs:
		for _, e := range x.Unwrap() {
			if !isCleanupSafe(e) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
