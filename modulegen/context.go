package main

import (
	"os"
	"path/filepath"
	"sort"
)

type Context struct {
	RootDir string
}

func (ctx *Context) DependabotConfigFile() string {
	return filepath.Join(ctx.GithubDir(), "dependabot.yml")
}

func (ctx *Context) DocsDir() string {
	return filepath.Join(ctx.RootDir, "docs")
}

func (ctx *Context) GithubDir() string {
	return filepath.Join(ctx.RootDir, ".github")
}

func (ctx *Context) GithubWorkflowsDir() string {
	return filepath.Join(ctx.GithubDir(), "workflows")
}

func (ctx *Context) getModulesByBaseDir(baseDir string) ([]string, error) {
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

func (ctx *Context) getMardownsByBaseDir(baseDir string) ([]string, error) {
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

func (ctx *Context) GetExamples() ([]string, error) {
	return ctx.getModulesByBaseDir("examples")
}

func (ctx *Context) GetModules() ([]string, error) {
	return ctx.getModulesByBaseDir("modules")
}

func (ctx *Context) GetExamplesDocs() ([]string, error) {
	return ctx.getMardownsByBaseDir("examples")
}

func (ctx *Context) GetModulesDocs() ([]string, error) {
	return ctx.getMardownsByBaseDir("modules")
}

func (ctx *Context) MkdocsConfigFile() string {
	return filepath.Join(ctx.RootDir, "mkdocs.yml")
}

func NewContext(dir string) *Context {
	return &Context{RootDir: dir}
}
