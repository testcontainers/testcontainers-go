package release

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewReleaseManager(t *testing.T) {
	testCases := []struct {
		name     string
		branch   string
		bumpType string
		dryRun   bool
	}{
		{
			name:     "main branch, minor bump, dry run",
			branch:   "main",
			bumpType: "minor",
			dryRun:   true,
		},
		{
			name:     "main branch, minor bump, no dry run",
			branch:   "main",
			bumpType: "minor",
			dryRun:   false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			r := NewReleaseManager(tc.branch, tc.bumpType, tc.dryRun)
			require.NotNil(tt, r)

			if tc.dryRun {
				require.IsType(tt, &dryRunReleaseManager{}, r)
			} else {
				require.IsType(tt, &releaseManager{}, r)
			}
		})
	}
}
