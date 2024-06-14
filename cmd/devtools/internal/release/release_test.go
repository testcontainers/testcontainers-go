package release

import (
	"fmt"
	"path/filepath"
	"strings"
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
		bumpType string
		// the version we are going to put in the version.go file after the release
		expectedVersion string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "Test Major bump with Dry Run",
			args: args{
				bumpType:        "major",
				expectedVersion: "1.0.0",
			},
		},
		{
			name: "Test Major bump without Dry Run",
			args: args{
				bumpType:        "major",
				expectedVersion: "1.0.0",
			},
		},
		{
			name: "Test Minor bump with Dry Run",
			args: args{
				bumpType:        "minor",
				expectedVersion: "0.2.0",
			},
		},
		{
			name: "Test Minor bump without Dry Run",
			args: args{
				bumpType:        "minor",
				expectedVersion: "0.2.0",
			},
		},
		{
			name: "Test Patch bump with Dry Run",
			args: args{
				bumpType:        "patch",
				expectedVersion: "0.1.1",
			},
		},
		{
			name: "Test Patch bump without Dry Run",
			args: args{
				bumpType:        "patch",
				expectedVersion: "0.1.1",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			//tt.Parallel()

			ctx := context.New(tt.TempDir())

			// create the releaser without dry-run, to perform the git operations in the temp directory
			dryRun := false
			releaser := NewTestReleaser(dryRun, ctx.RootDir, tc.args.bumpType)

			// we perform the bump from this current version
			initVersion := "0.0.1"
			// the current next development version in the version.go file,
			// which will receive the next development version after the bump
			nextDevelopmentVersion := "0.1.0"
			vNextDevelopmentVersion := fmt.Sprintf("v%s", nextDevelopmentVersion)

			// initialise project files
			initialiseProject(tt, ctx, rootCtx, initVersion, nextDevelopmentVersion)

			// init the git repository for testing
			gitClient := git.New(ctx, releaser.branch, false)
			if err := gitClient.InitRepository(); err != nil {
				tt.Fatalf("Error initializing git repository: %v", err)
			}

			// we need to force the pre-run so that the version file is created
			if err := releaser.PreRun(ctx); err != nil {
				tt.Fatalf("Pre() error = %v", err)
			}

			if err := releaser.Run(ctx); err != nil {
				tt.Errorf("Run() error = %v", err)
			}

			// because we are using a test release manager, the skipRemoteOps is set to true
			if !dryRun {
				// assert the commits has been produced
				output, err := gitClient.Log()
				if err != nil {
					tt.Fatalf("Error getting git log: %v", err)
				}

				if !strings.Contains(output, fmt.Sprintf("chore: use new version (%s) in modules and examples", vNextDevelopmentVersion)) {
					tt.Errorf("Expected new version commit message not found: %s", output)
				}
				if !strings.Contains(output, fmt.Sprintf("chore: prepare for next %s development version cycle (%s)", tc.args.bumpType, tc.args.expectedVersion)) {
					tt.Errorf("Expected next development version commit message not found: %s", output)
				}

				// assert the tags for the library and all the modules exist
				output, err = gitClient.ListTags()
				if err != nil {
					tt.Fatalf("Error listing git tags: %v", err)
				}

				if !strings.Contains(output, vNextDevelopmentVersion) {
					tt.Errorf("Expected core version tag not found: %s", output)
				}
				for _, module := range modules {
					moduleTag := fmt.Sprintf("%s/%s/%s", "modules", module, vNextDevelopmentVersion)
					if !strings.Contains(output, moduleTag) {
						tt.Errorf("Expected module version tag not found: %s", output)
					}
				}
				for _, example := range examples {
					exampleTag := fmt.Sprintf("%s/%s/%s", "examples", example, vNextDevelopmentVersion)
					if !strings.Contains(output, exampleTag) {
						tt.Errorf("Expected example version tag not found: %s", output)
					}
				}

				// assert the next development version has been applied to the version.go file
			}
		})
	}
}
