package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	if err != nil {
		t.Fatal(err)
	}

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

			// init the git repository for testing
			gitClient := git.New(ctx, releaser.branch, tc.args.dryRun)
			if err := gitClient.InitRepository(); err != nil {
				tt.Fatalf("Error initializing git repository: %v", err)
			}

			if err := releaser.PreRun(ctx, gitClient); err != nil {
				tt.Fatalf("Pre() error = %v", err)
			}

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
		})
	}
}

func assertBumpFiles(t *testing.T, ctx context.Context, version string) {
	// mkdocs.yml
	read, err := os.ReadFile(filepath.Join(ctx.RootDir, bumpFiles[0]))
	if err != nil {
		t.Fatal(err)
	}

	if string(read) != "extra:\n  latest_version: v"+version {
		t.Errorf("Expected extra:\n  latest_version: v%s, got %s", version, string(read))
	}

	// sonar-project.properties file
	read, err = os.ReadFile(filepath.Join(ctx.RootDir, bumpFiles[1]))
	if err != nil {
		t.Fatal(err)
	}

	if string(read) != "sonar.projectVersion=v"+version {
		t.Errorf("Expected sonar.projectVersion=v%s, got %s", version, string(read))
	}
}

func assertMarkdownFiles(t *testing.T, ctx context.Context, expected string) {
	for _, f := range testMarkdownFiles {
		read, err := os.ReadFile(filepath.Join(ctx.DocsDir(), f))
		if err != nil {
			t.Fatal(err)
		}

		if string(read) != expected {
			t.Errorf("Expected %s, got %s", expected, string(read))
		}
	}
}

func assertModules(t *testing.T, ctx context.Context, isModule bool, version string) {
	types := modules

	moduleType := "modules"
	if !isModule {
		moduleType = "examples"
		types = examples
	}

	for _, m := range types {
		read, err := os.ReadFile(filepath.Join(ctx.RootDir, moduleType, m, "go.mod"))
		if err != nil {
			t.Fatal(err)
		}

		content := string(read)

		expected := `module github.com/testcontainers/testcontainers-go/` + moduleType + `/` + m + `

go 1.21`

		if !strings.HasPrefix(content, expected) {
			t.Errorf("Expected %s, got %s", expected, content)
		}

		expecteds := []string{
			"require github.com/testcontainers/testcontainers-go v" + version,
			"replace github.com/testcontainers/testcontainers-go => ../..",
		}

		for _, e := range expecteds {
			if !strings.Contains(content, e) {
				t.Errorf("Expected %s, got %s", e, content)
			}
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
			if err != nil {
				tt.Fatal(err)
			}

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
				if err != nil {
					tt.Fatal(err)
				}

				if string(read) != expected[f] {
					tt.Errorf("Expected %s, got %s", expected[f], string(read))
				}
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tmp := t.TempDir()

	ctx := context.New(tmp)

	createVersionFile(t, ctx, "1.2.3")

	version, err := extractCurrentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got %s", version)
	}
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
			if err != nil {
				tt.Fatal(err)
			}

			err = replaceInFile(tc.dryRun, tc.regex, file, "latest_version: .*", "latest_version: "+tc.new)
			if err != nil {
				tt.Fatal(err)
			}

			read, err := os.ReadFile(file)
			if err != nil {
				tt.Fatal(err)
			}

			if string(read) != "latest_version: "+tc.expected {
				tt.Errorf("Expected latest_version: %s, got %s", string(read), tc.expected)
			}
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
			if err != nil {
				tt.Fatal(err)
			}

			expected := map[string]string{}
			for _, f := range testMarkdownFiles {
				expected[f] = tc.expected
			}

			for f, content := range expected {
				read, err := os.ReadFile(filepath.Join(ctx.DocsDir(), f))
				if err != nil {
					tt.Fatal(err)
				}

				if string(read) != content {
					tt.Errorf("Expected %s, got %s", tc.expected, string(read))
				}
			}
		})
	}
}
