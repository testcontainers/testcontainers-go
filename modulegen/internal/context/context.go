package context

import (
	"os"
	"path/filepath"
	"sort"
)

// Context is the context for the module generation.
// It provides the necessary information to generate the module or example,
// such as the different paths to the files and directories.
type Context struct {
	RootDir string
}

// DependabotConfigFile returns, from the root directory, the relative path
// to the dependabot config file, "/.github/dependabot.yml".
func (ctx Context) DependabotConfigFile() string {
	return filepath.Join(ctx.GithubDir(), "dependabot.yml")
}

// DocsDir returns, from the root directory, the relative path to the docs directory, "/docs".
func (ctx Context) DocsDir() string {
	return filepath.Join(ctx.RootDir, "docs")
}

// ExamplesDir returns, from the root directory, the relative path to the examples directory, "/examples".
func (ctx Context) ExamplesDir() string {
	return filepath.Join(ctx.RootDir, "examples")
}

// ExamplesDocsDir returns, from the docs directory, the relative path to the examples directory, "/docs/examples".
func (ctx Context) ExamplesDocsDir() string {
	return filepath.Join(ctx.DocsDir(), "examples")
}

// ModulesDir returns, from the root directory, the relative path to the modules directory, "/modules".
func (ctx Context) ModulesDir() string {
	return filepath.Join(ctx.RootDir, "modules")
}

// ModulesDocsDir returns, from the docs directory, the relative path to the modules directory, "/docs/modules".
func (ctx Context) ModulesDocsDir() string {
	return filepath.Join(ctx.DocsDir(), "modules")
}

// GithubDir returns, from the root directory, the relative path to the github directory, "/.github".
func (ctx Context) GithubDir() string {
	return filepath.Join(ctx.RootDir, ".github")
}

// GithubWorkflowsDir returns, from the github directory, the relative path to the workflows directory, "/.github/workflows".
func (ctx Context) GithubWorkflowsDir() string {
	return filepath.Join(ctx.GithubDir(), "workflows")
}

// GoModFile returns, from the root directory, the relative path to the go.mod file, "/go.mod".
func (ctx Context) GoModFile() string {
	return filepath.Join(ctx.RootDir, "go.mod")
}

// getModulesByBaseDir returns, from the root directory, the relative paths to the modules or examples,
// depending on the base directory.
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

// getMarkdownsFromDir returns, from the docs directory, the relative paths to the markdown files,
// depending on the base directory.
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

// GetExamples returns, from the root directory, a slice with the names of the examples
// that are located in the "examples" directory.
func (ctx Context) GetExamples() ([]string, error) {
	return ctx.getModulesByBaseDir("examples")
}

// GetModules returns, from the root directory, a slice with the names of the modules
// that are located in the "modules" directory.
func (ctx Context) GetModules() ([]string, error) {
	return ctx.getModulesByBaseDir("modules")
}

// GetExamplesDocs returns, from the docs directory, a slice with the names of the markdown files
// that are located in the "docs/examples" directory.
func (ctx Context) GetExamplesDocs() ([]string, error) {
	return ctx.getMarkdownsFromDir("examples")
}

// GetModulesDocs returns, from the docs directory, a slice with the names of the markdown files
// that are located in the "docs/modules" directory.
func (ctx Context) GetModulesDocs() ([]string, error) {
	return ctx.getMarkdownsFromDir("modules")
}

// MkdocsConfigFile returns, from the root directory, the relative path to the mkdocs config file, "/mkdocs.yml".
func (ctx Context) MkdocsConfigFile() string {
	return filepath.Join(ctx.RootDir, "mkdocs.yml")
}

// VSCodeWorkspaceFile returns, from the root directory, the relative path to the vscode workspace file, "/.vscode/.testcontainers-go.code-workspace".
func (ctx Context) VSCodeWorkspaceFile() string {
	return filepath.Join(ctx.RootDir, ".vscode", ".testcontainers-go.code-workspace")
}

// New returns a new Context with the given root directory.
func New(dir string) Context {
	return Context{RootDir: dir}
}

// GetRootContext returns a new Context with the current working directory.
func GetRootContext() (Context, error) {
	current, err := os.Getwd()
	if err != nil {
		return Context{}, err
	}
	return New(filepath.Dir(current)), nil
}
