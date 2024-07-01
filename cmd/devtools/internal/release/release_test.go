package release

import (
	gocontext "context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	wiremock "github.com/wiremock/wiremock-testcontainers-go"

	"github.com/testcontainers/testcontainers-go"
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
			tt.Parallel()

			logConsumer := &wiremockLogConsumer{}

			mockProxyContainer := startGolangProxy(t, logConsumer)
			tt.Cleanup(func() {
				if err := mockProxyContainer.Terminate(gocontext.Background()); err != nil {
					tt.Fatalf("Error terminating container: %v", err)
				}
			})

			mockProxyURL, err := mockProxyContainer.PortEndpoint(gocontext.Background(), "8080", "http")
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.New(tt.TempDir())

			// create the releaser without dry-run, to perform the git operations in the temp directory
			dryRun := false
			releaser := NewTestReleaser(dryRun, ctx.RootDir, tc.args.bumpType, mockProxyURL)

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
				tt.Fatalf("Run() error = %v", err)
			}

			// wait for the log consumer to process all the logs
			for i := 0; i < 10; i++ {
				// 3 examples, 3 modules and the core
				if len(logConsumer.lines) == 7 {
					break
				}
				time.Sleep(50 * time.Millisecond)
			}
			// 3 examples, 3 modules and the core
			if len(logConsumer.lines) != 7 {
				tt.Errorf("Expected 7 hits to the golang proxy, got %d", len(logConsumer.lines))
			}

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
			version, err := extractCurrentVersion(ctx)
			if err != nil {
				tt.Fatalf("Error extracting current version: %v", err)
			}

			if version != tc.args.expectedVersion {
				tt.Errorf("Expected next development version not found: %s", version)
			}
		})
	}
}

// wiremockLogConsumer is a LogConsumer for wiremock, filtering the logs to the requests we are interested in
type wiremockLogConsumer struct {
	lines []string
}

// Accept prints the log to stdout
func (lc *wiremockLogConsumer) Accept(l testcontainers.Log) {
	lines := strings.Split(string(l.Content), "\n")

	for _, line := range lines {
		if strings.Contains(line, "GET /github.com/testcontainers/testcontainers-go") {
			lc.lines = append(lc.lines, line)
		}
	}
}

func startGolangProxy(t *testing.T, consumer testcontainers.LogConsumer) *wiremock.WireMockContainer {
	goCtx := gocontext.Background()

	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithImage("wiremock/wiremock:3.8.0"),
		testcontainers.WithEnv(map[string]string{
			// enable verbose mode in order to capture the requests in the log consumer
			"WIREMOCK_OPTIONS": "--verbose",
		}),
		testcontainers.WithLogConsumers(consumer),
		wiremock.WithMappingFile("proxy", filepath.Join("testdata", "proxy.json")),
	}

	mockProxyContainer, err := wiremock.RunContainer(goCtx, opts...)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	return mockProxyContainer
}
