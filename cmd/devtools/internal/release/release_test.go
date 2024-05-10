package release

import (
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
)

func TestGetSemverVersion(t *testing.T) {
	type args struct {
		bumpType string
		vVersion string
	}
	testCases := []struct {
		name    string
		args    args
		want    string // the expected version must not include the 'v' prefix
		wantErr bool
	}{
		{
			name: "major bump",
			args: args{
				bumpType: "major",
				vVersion: "v1.0.0",
			},
			want:    "2.0.0",
			wantErr: false,
		},
		{
			name: "minor bump",
			args: args{
				bumpType: "minor",
				vVersion: "v1.0.0",
			},
			want:    "1.1.0",
			wantErr: false,
		},
		{
			name: "patch bump",
			args: args{
				bumpType: "patch",
				vVersion: "1.0.0",
			},
			want:    "1.0.1",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			got, err := getSemverVersion(tc.args.bumpType, tc.args.vVersion)
			if (err != nil) != tc.wantErr {
				tt.Errorf("getSemverVersion() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.want {
				tt.Errorf("getSemverVersion() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	// uses two directories up to get the root directory
	rootCtx, err := context.GetRootContext()
	if err != nil {
		t.Fatal(err)
	}

	// we need to go two directories up more to get the root directory
	rootCtx = context.New(filepath.Dir(filepath.Dir(rootCtx.RootDir)))

	type args struct {
		dryRun   bool
		bumpType string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "Test Major bump with Dry Run",
			args: args{
				dryRun:   true,
				bumpType: "major",
			},
		},
		{
			name: "Test Major bump without Dry Run",
			args: args{
				dryRun:   false,
				bumpType: "major",
			},
		},
		{
			name: "Test Minor bump with Dry Run",
			args: args{
				dryRun:   true,
				bumpType: "minor",
			},
		},
		{
			name: "Test Minor bump without Dry Run",
			args: args{
				dryRun:   false,
				bumpType: "minor",
			},
		},
		{
			name: "Test Patch bump with Dry Run",
			args: args{
				dryRun:   true,
				bumpType: "patch",
			},
		},
		{
			name: "Test Patch bump without Dry Run",
			args: args{
				dryRun:   false,
				bumpType: "patch",
			},
		},
	}

	for _, tc := range testCases {
		if tc.args.dryRun {
			continue
		}

		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			// tt.Parallel()

			ctx := context.New(tt.TempDir())

			releaser := TestReleaser{
				dryRun:   tc.args.dryRun,
				branch:   "main-" + filepath.Base(filepath.Dir(ctx.RootDir)),
				bumpType: tc.args.bumpType,
			}

			initVersion := "0.0.1"
			nextVersion := "0.1.0"

			// initialise project files
			initialiseProject(tt, ctx, rootCtx, initVersion, nextVersion)

			// init the git repository for testing
			gitClient := git.New(ctx, releaser.branch, tc.args.dryRun)
			if err := gitClient.InitRepository(); err != nil {
				tt.Fatalf("Error initializing git repository: %v", err)
			}

			if !tc.args.dryRun {
				// we need to force the pre-run so that the version file is created
				if err := releaser.PreRun(ctx); err != nil {
					tt.Fatalf("Pre() error = %v", err)
				}
			}

			if err := releaser.Run(ctx); err != nil {
				tt.Errorf("Run() error = %v", err)
			}

			if !tc.args.dryRun {
				// assert the commit has been produced
				// assert the tags for the library and all the modules exist
				// assert the next development version has been applied to the version.go file
			}
		})
	}
}
