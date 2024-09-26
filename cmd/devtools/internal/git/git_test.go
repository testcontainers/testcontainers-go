package git

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

func TestCheckOriginRemote(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "Test Has Origin with Dry Run",
			dryRun: true,
		},
		{
			name:   "Test Has Origin without Dry Run",
			dryRun: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.New(tt.TempDir())

			gitClient := New(ctx, "main", tc.dryRun)
			err := gitClient.InitRepository("foo")
			require.NoError(tt, err)

			cleanUp, err := gitClient.CheckOriginRemote()
			require.NoError(tt, err)

			tt.Cleanup(func() {
				require.NoError(tt, cleanUp())
			})

			if !tc.dryRun {
				remotes, err := gitClient.Remotes()
				require.NoError(tt, err)
				require.Len(tt, remotes, 4)

				// verify that the origin remote contains the Github repository URL
				if r, ok := remotes["origin-(fetch)"]; !ok || r != tcOrigin {
					tt.Fatalf("Expected origin-fetch remote to be %s, got %s", tcOrigin, r)
				}
				if r, ok := remotes["origin-(push)"]; !ok || r != tcOrigin {
					tt.Fatalf("Expected origin-fetch remote to be %s, got %s", tcOrigin, r)
				}
			}
		})
	}
}
