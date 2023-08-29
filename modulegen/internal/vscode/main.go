package vscode

import (
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(rootDir string, examples []string, modules []string) error {
	projectDirectories := newProjectDirectories(rootDir, examples, modules)
	name := "testcontainers-go.code-workspace.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(rootDir, ".vscode", ".testcontainers-go.code-workspace")

	return internal_template.GenerateFile(t, exampleFilePath, name, projectDirectories)
}
