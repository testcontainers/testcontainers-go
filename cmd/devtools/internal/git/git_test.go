package git

import (
	"testing"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
)

func TestHasOriginRemote(t *testing.T) {
	tests := []struct {
		name    string
		dryRun  bool
		wantErr bool
	}{
		{
			name:    "Test Has Origin with Dry Run",
			dryRun:  true,
			wantErr: false,
		},
		{
			name:    "Test Has Origin without Dry Run",
			dryRun:  false,
			wantErr: false,
		},
		{
			name:    "Test Has no Origin without Dry Run",
			dryRun:  false,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.New(tt.TempDir())

			gitClient := New(ctx, "main", tc.dryRun)
			if err := gitClient.InitRepository(); err != nil {
				tt.Fatalf("Error initializing git repository: %v", err)
			}

			if tc.wantErr {
				// updating the origin for the error case
				gitClient.origin = "foo"
			}

			err := gitClient.HasOriginRemote()
			if (err != nil) != tc.wantErr {
				tt.Errorf("HasOriginRemote() error = %v, wantErr %v", err, tc.wantErr)
			} else if err == nil && tc.wantErr {
				tt.Errorf("HasOriginRemote() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
