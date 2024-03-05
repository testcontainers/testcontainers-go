package context

import (
	"os"
	"path/filepath"
	"sort"
)

type Context struct {
	RootDir string
}

func (ctx Context) DocsDir() string {
	return filepath.Join(ctx.RootDir, "docs")
}

func (ctx Context) GithubDir() string {
	return filepath.Join(ctx.RootDir, ".github")
}

func (ctx Context) GithubWorkflowsDir() string {
	return filepath.Join(ctx.GithubDir(), "workflows")
}

func (ctx Context) GoModFile() string {
	return filepath.Join(ctx.RootDir, "go.mod")
}

func (ctx Context) getModulesByBaseDir(baseDir string) ([]string, error) {
	dir := filepath.Join(ctx.RootDir, baseDir)

	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)

	for _, f := range allFiles {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func (ctx Context) getMarkdownsFromDir(baseDir string) ([]string, error) {
	dir := filepath.Join(ctx.DocsDir(), baseDir)

	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)

	for _, f := range allFiles {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".md" {
			dirs = append(dirs, f.Name())
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

func (ctx Context) GetExamples() ([]string, error) {
	return ctx.getModulesByBaseDir("examples")
}

func (ctx Context) GetModules() ([]string, error) {
	return ctx.getModulesByBaseDir("modules")
}

func (ctx Context) GetExamplesDocs() ([]string, error) {
	return ctx.getMarkdownsFromDir("examples")
}

func (ctx Context) GetModulesDocs() ([]string, error) {
	return ctx.getMarkdownsFromDir("modules")
}

func (ctx Context) MkdocsConfigFile() string {
	return filepath.Join(ctx.RootDir, "mkdocs.yml")
}

func (ctx Context) SonarProjectFile() string {
	return filepath.Join(ctx.RootDir, "sonar-project.properties")
}

func (ctx Context) VSCodeWorkspaceFile() string {
	return filepath.Join(ctx.RootDir, ".vscode", ".testcontainers-go.code-workspace")
}

func New(dir string) Context {
	return Context{RootDir: dir}
}

func GetRootContext() (Context, error) {
	current, err := os.Getwd()
	if err != nil {
		return Context{}, err
	}
	return New(filepath.Dir(current)), nil
}
