package testcontainers

import (
	"errors"
	"fmt"
	"testing"

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
			want: false,
		},
		"warp": {
			err:  fmt.Errorf("wrap: %w", notFoundError{}),
			want: true,
		},
		"multi-warp": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w", notFoundError{})),
			want: true,
		},
		"multi-warp-other": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w", errors.New("other"))),
			want: false,
		},
		"multi-warp-other-not-found": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w %w", errors.New("other"), notFoundError{})),
			want: false,
		},
		"multi-warp-not-found-nil": {
			err:  fmt.Errorf("wrap: %w", fmt.Errorf("wrap: %w %w", nil, notFoundError{})),
			want: true,
		},
		"multi-join-not-found-other": {
			err:  errors.Join(nil, fmt.Errorf("wrap: %w", errors.Join(notFoundError{}, errors.New("other")))),
			want: false,
		},
		"multi-join-not-found-nil": {
			err:  errors.Join(nil, fmt.Errorf("wrap: %w", errors.Join(notFoundError{}, nil))),
			want: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.want, isCleanupSafe(tc.err))
		})
	}
}
