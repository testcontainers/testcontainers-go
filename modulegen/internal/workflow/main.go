package workflow

import (
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(githubWorkflowsDir string, examples []string, modules []string) error {
	projectDirectories := newProjectDirectories(examples, modules)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(githubWorkflowsDir, "ci.yml")

	return internal_template.GenerateFile(t, exampleFilePath, name, projectDirectories)
}
