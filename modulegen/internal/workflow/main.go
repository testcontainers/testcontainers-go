package workflow

import (
	"path/filepath"
	"text/template"

	internal "github.com/testcontainers/testcontainers-go/modulegen/internal"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(githubWorkflowsDir string, examples []string, modules []string) error {
	projectDirectories := internal.NewProjectDirectories(examples, modules)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}
	return internal_template.Generate(t, filepath.Join(githubWorkflowsDir, "ci.yml"), name, projectDirectories)
}
