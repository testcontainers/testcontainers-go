package workflows

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	templates "github.com/testcontainers/testcontainers-go/cmd/pkg/templates"
)

func GenerateFromContext(ctx *Context) error {
	projectDirectories := newProjectDirectories(ctx)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_templates", "workflows", name))
	if err != nil {
		return err
	}
	return templates.Generate(t, ctx.ciFile(), name, projectDirectories)
}

func newProjectDirectories(ctx *Context) *ProjectDirectories {
	return &ProjectDirectories{
		Examples: strings.Join(getProjectNames(ctx, "examples"), ", "),
		Modules:  strings.Join(getProjectNames(ctx, "modules"), ", "),
	}
}

func getProjectNames(ctx *Context, baseDir string) []string {
	dirs := make([]string, 0)
	dir := filepath.Join(ctx.RootContext.RootDir, baseDir)

	allFiles, err := os.ReadDir(dir)
	if err != nil {
		return dirs
	}

	for _, f := range allFiles {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}
