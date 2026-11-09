package testcontainers

import (
	"errors"
	"fmt"
	"testing"

	"github.com/containerd/errdefs"
	"github.com/stretchr/testify/require"
)

func ExampleSkipIfProviderIsNotHealthy() {
	SkipIfProviderIsNotHealthy(&testing.T{})
}

type notFoundError struct{}

func (notFoundError) NotFound() {}

func (notFoundError) Error() string {
	return "not found"
}

// errPodmanRemovedDuringInspect is the error Podman returns when a container
// is removed, e.g. automatically once it exits, while it is being inspected.
// The message always includes the container's ID, which can be any.
var errPodmanRemovedDuringInspect = errdefs.ErrInternal.WithMessage("Error response from daemon: no container with ID abc123 found in database: no such container")

func Test_isNotFound(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"nil": {
			err:  nil,
			want: true,
		},
		"join-nils": {
			err:  errors.Join(nil, nil),
			want: true,
		},
		"join-nil-not-found": {
			err:  errors.Join(nil, notFoundError{}),
			want: true,
		},
		"not-found": {
			err:  notFoundError{},
			want: true,
		},
		"other": {
			err:  errors.New("other"),
			want: false,
		},
		"join-other": {
			err:  errors.Join(nil, notFoundError{}, errors.New("other")),
			want: true,
		},
		"wrap": {
			err:  fmt.Errorf("wrap: %w", notFoundError{}),
			want: true,
		},
		"multi-wrap": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w", notFoundError{})),
			want: true,
		},
		"multi-wrap-other": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w", errors.New("other"))),
			want: false,
		},
		"multi-wrap-other-not-found": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w %w", errors.New("other"), notFoundError{})),
			want: true,
		},
		"multi-wrap-not-found-nil": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w %w", nil, notFoundError{})),
			want: true,
		},
		"multi-join-not-found-other": {
			err:  errors.Join(nil, fmt.Errorf("wrap: %w", errors.Join(notFoundError{}, errors.New("other")))),
			want: true,
		},
		"multi-join-not-found-nil": {
			err:  errors.Join(nil, fmt.Errorf("wrap: %w", errors.Join(notFoundError{}, nil))),
			want: true,
		},
		"podman-removed-during-inspect": {
			err:  fmt.Errorf("container state: %w", errPodmanRemovedDuringInspect),
			want: true,
		},
		"internal-other": {
			err:  errdefs.ErrInternal.WithMessage("database is locked"),
			want: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, isCleanupSafe(tc.err))
		})
	}
}
