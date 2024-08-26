package git

import (
	"testing"

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
			if err := gitClient.InitRepository("foo"); err != nil {
				tt.Fatalf("Error initializing git repository: %v", err)
			}

			cleanUp, err := gitClient.CheckOriginRemote()
			if err != nil {
				tt.Fatalf("Error checking origin remote: %v", err)
			}
			tt.Cleanup(func() {
				if err := cleanUp(); err != nil {
					tt.Fatalf("Error cleaning up: %v", err)
				}
			})

			if !tc.dryRun {
				remotes, err := gitClient.Remotes()
				if err != nil {
					tt.Fatalf("Error getting remotes: %v", err)
				}

				if len(remotes) != 4 {
					tt.Errorf("Expected 4 remotes, got %d", len(remotes))
				}

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
