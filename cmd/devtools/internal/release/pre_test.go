package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/git"
)

var (
	bumpFiles         = []string{"mkdocs.yml", "sonar-project.properties"}
	testMarkdownFiles = []string{"file1.md", "file2.md", "file3.md"}
	modules           = []string{"module1", "module2", "module3"}
	examples          = []string{"example1", "example2", "example3"}
)

func TestPre(t *testing.T) {
	t.Parallel()

	// uses two directories up to get the root directory
	rootCtx, err := context.GetRootContext()
	require.NoError(t, err)

	// we need to go two directories up more to get the root directory
	rootCtx = context.New(filepath.Dir(filepath.Dir(rootCtx.RootDir)))

	type args struct {
		dryRun bool
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "Test Pre with Dry Run",
			args: args{
				dryRun: true,
			},
		},
		{
			name: "Test Pre without Dry Run",
			args: args{
				dryRun: false,
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			ctx := context.New(tt.TempDir())

			releaser := NewTestReleaser(tc.args.dryRun, ctx.RootDir, "minor", "")

			initVersion := "0.0.1"
			nextVersion := "0.1.0"

			// initialise project files
			initialiseProject(tt, ctx, rootCtx, initVersion, nextVersion)

			expectedRemote := "foo"

			// init the git repository for testing
			gitClient := git.New(ctx, releaser.branch, tc.args.dryRun)
			err := gitClient.InitRepository(expectedRemote)
			require.NoError(tt, err)

			err = releaser.PreRun(ctx, gitClient)
			require.NoError(tt, err)

			expectedVersion := nextVersion
			expectedMarkDown := sinceVersionText(nextVersion)
			if tc.args.dryRun {
				expectedVersion = initVersion
				expectedMarkDown = nonReleasedText
			}

			assertBumpFiles(tt, ctx, expectedVersion)
			assertModules(tt, ctx, true, expectedVersion)
			assertModules(tt, ctx, false, expectedVersion)
			assertMarkdownFiles(tt, ctx, expectedMarkDown)

			if !tc.args.dryRun {
				assertGitState(tt, gitClient, expectedRemote)
			}
		})
	}
}

func assertBumpFiles(t *testing.T, ctx context.Context, version string) {
	t.Helper()

	// mkdocs.yml
	read, err := os.ReadFile(filepath.Join(ctx.RootDir, bumpFiles[0]))
	require.NoError(t, err)
	require.Equal(t, "extra:\n  latest_version: v"+version, string(read))

	// sonar-project.properties file
	read, err = os.ReadFile(filepath.Join(ctx.RootDir, bumpFiles[1]))
	require.NoError(t, err)
	require.Equal(t, "sonar.projectVersion=v"+version, string(read))
}

func assertGitState(t *testing.T, gitClient *git.GitClient, expectedRemote string) {
	t.Helper()

	remotes, err := gitClient.Remotes()
	require.NoError(t, err)
	require.Len(t, remotes, 2)

	// verify that the origin remote contains the expected Github repository URL
	if r, ok := remotes["origin-(fetch)"]; !ok || r != expectedRemote {
		t.Fatalf("Expected origin-fetch remote to be %s, got %s", expectedRemote, r)
	}
	if r, ok := remotes["origin-(push)"]; !ok || r != expectedRemote {
		t.Fatalf("Expected origin-fetch remote to be %s, got %s", expectedRemote, r)
	}
}

func assertMarkdownFiles(t *testing.T, ctx context.Context, expected string) {
	t.Helper()

	for _, f := range testMarkdownFiles {
		read, err := os.ReadFile(filepath.Join(ctx.DocsDir(), f))
		require.NoError(t, err)
		require.Equal(t, expected, string(read))
	}
}

func assertModules(t *testing.T, ctx context.Context, isModule bool, version string) {
	t.Helper()

	types := modules

	moduleType := "modules"
	if !isModule {
		moduleType = "examples"
		types = examples
	}

	for _, m := range types {
		read, err := os.ReadFile(filepath.Join(ctx.RootDir, moduleType, m, "go.mod"))
		require.NoError(t, err)

		content := string(read)

		expected := `module github.com/testcontainers/testcontainers-go/` + moduleType + `/` + m + `

go 1.21`

		require.True(t, strings.HasPrefix(content, expected))

		expecteds := []string{
			"require github.com/testcontainers/testcontainers-go v" + version,
			"replace github.com/testcontainers/testcontainers-go => ../..",
		}

		for _, e := range expecteds {
			require.Contains(t, content, e)
		}
	}
}

func TestBumpVersion(t *testing.T) {
	const version = "1.2.3"
	newVersion := "4.5.6"

	testCases := []struct {
		name   string
		dryRun bool
	}{
		{
			name:   "Test Bump With Dry Run",
			dryRun: true,
		},
		{
			name:   "Test Bump Without Dry Run",
			dryRun: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			tmp := tt.TempDir()

			ctx := context.New(tmp)

			createBumpFiles(tt, ctx, version)

			err := bumpVersion(ctx, tc.dryRun, "v"+newVersion)
			require.NoError(tt, err)

			var expected map[string]string
			// it's important to note that the YAML files use two spaces as indentation
			if tc.dryRun {
				expected = map[string]string{
					bumpFiles[0]: `extra:
  latest_version: v` + version,
					bumpFiles[1]: "sonar.projectVersion=v" + version,
				}
			} else {
				expected = map[string]string{
					bumpFiles[0]: `extra:
  latest_version: v` + newVersion,
					bumpFiles[1]: "sonar.projectVersion=v" + newVersion,
				}
			}

			for f := range expected {
				read, err := os.ReadFile(filepath.Join(ctx.RootDir, f))
				require.NoError(tt, err)
				require.Equal(tt, expected[f], string(read))
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tmp := t.TempDir()

	ctx := context.New(tmp)

	createVersionFile(t, ctx, "1.2.3")

	version, err := extractCurrentVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.2.3", version)
}

func TestReplaceInFile(t *testing.T) {
	const defaultVersion = "1.2.3"

	testCases := []struct {
		name     string
		dryRun   bool
		regex    bool
		new      string
		expected string
	}{
		{
			name:     "Test Replace Using Regex Without Dry Run",
			dryRun:   false,
			regex:    true,
			new:      "4.5.6",
			expected: "4.5.6",
		},
		{
			name:     "Test Replace Using Regex With Dry Run",
			dryRun:   true,
			regex:    true,
			new:      "4.5.6",
			expected: defaultVersion,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			tmp := tt.TempDir()
			file := filepath.Join(tmp, "file.txt")

			content := "latest_version: " + defaultVersion

			err := os.WriteFile(file, []byte(content), 0o644)
			require.NoError(tt, err)

			err = replaceInFile(tc.dryRun, tc.regex, file, "latest_version: .*", "latest_version: "+tc.new)
			require.NoError(tt, err)

			read, err := os.ReadFile(file)
			require.NoError(tt, err)
			require.Equal(tt, "latest_version: "+tc.expected, string(read))
		})
	}
}

func TestProcessMarkdownFiles(t *testing.T) {
	const version = "1.2.3"
	releasedText := sinceVersionText(version)

	testCases := []struct {
		name     string
		dryRun   bool
		expected string
	}{
		{
			name:     "Test Process With Dry Run",
			dryRun:   true,
			expected: nonReleasedText,
		},
		{
			name:     "Test Process Without Dry Run",
			dryRun:   false,
			expected: releasedText,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(tt *testing.T) {
			tt.Parallel()

			tmp := tt.TempDir()

			ctx := context.New(tmp)

			createMarkdownFiles(tt, ctx)

			err := processMarkdownFiles(tc.dryRun, ctx.DocsDir(), version)
			require.NoError(tt, err)

			expected := map[string]string{}
			for _, f := range testMarkdownFiles {
				expected[f] = tc.expected
			}

			for f, content := range expected {
				read, err := os.ReadFile(filepath.Join(ctx.DocsDir(), f))
				require.NoError(tt, err)
				require.Equal(tt, content, string(read))
			}
		})
	}
}
