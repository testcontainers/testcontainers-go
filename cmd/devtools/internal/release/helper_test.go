package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/testcontainers/testcontainers-go/devtools/internal/context"
	"github.com/testcontainers/testcontainers-go/devtools/internal/module"
)

type testReleaser struct {
	dryRun        bool
	branch        string
	bumpType      string
	proxyURL      string
	skipRemoteOps bool
}

func NewTestReleaser(dryRun bool, rootDir string, bumpType string, proxyURL string) *testReleaser {
	return &testReleaser{
		dryRun:        dryRun,
		branch:        "main-" + filepath.Base(filepath.Dir(rootDir)),
		bumpType:      bumpType,
		skipRemoteOps: true,
		proxyURL:      proxyURL,
	}
}

func (p *testReleaser) PreRun(ctx context.Context) error {
	return preRun(ctx, p.branch, p.dryRun)
}

func (p *testReleaser) Run(ctx context.Context) error {
	return run(ctx, p.branch, p.bumpType, p.dryRun, p.skipRemoteOps, p.proxyURL)
}

func createBumpFiles(t *testing.T, ctx context.Context, version string) {
	files := map[string]string{
		bumpFiles[0]: `extra:
  latest_version: v` + version,
		bumpFiles[1]: "sonar.projectVersion=v" + version,
	}

	for f, content := range files {
		err := os.WriteFile(filepath.Join(ctx.RootDir, f), []byte(content), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createMarkdownFiles(t *testing.T, ctx context.Context) {
	docsDir := filepath.Join(ctx.RootDir, "docs")

	err := os.MkdirAll(docsDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range testMarkdownFiles {
		err = os.WriteFile(filepath.Join(docsDir, f), []byte(nonReleasedText), 0o644)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createModFile(t *testing.T, ctx context.Context) {
	content := `module github.com/testcontainers/testcontainers-go
go 1.21`

	err := os.WriteFile(filepath.Join(ctx.RootDir, "go.mod"), []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func createModules(t *testing.T, ctx context.Context, rootCtx context.Context, version string) {
	rootTemplatesDir := filepath.Join(rootCtx.RootDir, "cmd", "devtools", "_template")
	templatesDir := filepath.Join(ctx.RootDir, "cmd", "devtools")

	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("cp", "-R", rootTemplatesDir, templatesDir)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("cp", filepath.Join(rootCtx.RootDir, "commons-test.mk"), ctx.RootDir)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	modulesDir := filepath.Join(ctx.RootDir, "modules")
	if err := os.MkdirAll(modulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("cp", filepath.Join(rootCtx.RootDir, "modules", "Makefile"), modulesDir)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	mg := module.Generator{}
	for _, module := range modules {
		err := mg.AddModule(ctx, context.TestcontainersModule{
			IsModule:  true,
			Name:      module,
			Image:     module + ":latest",
			TitleName: cases.Title(language.Und, cases.NoLower).String(module),
			TCVersion: version,
			Context:   ctx,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	examplesDir := filepath.Join(ctx.RootDir, "examples")
	if err := os.MkdirAll(examplesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd = exec.Command("cp", filepath.Join(rootCtx.RootDir, "examples", "Makefile"), examplesDir)
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	for _, example := range examples {
		err := mg.AddModule(ctx, context.TestcontainersModule{
			IsModule:  false,
			Name:      example,
			Image:     example + ":latest",
			TitleName: cases.Title(language.Und, cases.NoLower).String(example),
			TCVersion: version,
			Context:   ctx,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func createVersionFile(t *testing.T, ctx context.Context, version string) {
	internalDir := filepath.Join(ctx.RootDir, "internal")
	content := `const Version = "` + version + `"`

	err := os.MkdirAll(internalDir, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(internalDir, "version.go"), []byte(content), 0o644)
	if err != nil {
		t.Fatal(err)
	}
}

func initialiseProject(t *testing.T, ctx context.Context, rootCtx context.Context, initVersion string, nextDevelopmentVersion string) {
	createVersionFile(t, ctx, nextDevelopmentVersion)
	createBumpFiles(t, ctx, initVersion)
	createMarkdownFiles(t, ctx)
	createModFile(t, ctx)
	createModules(t, ctx, rootCtx, initVersion)
}
