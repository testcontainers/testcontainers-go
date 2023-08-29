package vscode

import (
	"io"
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(wr io.Writer, examples []string, modules []string) error {
	projectDirectories := newProjectDirectories(examples, modules)
	name := "testcontainers-go.code-workspace.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	return internal_template.Generate(t, wr, name, projectDirectories)
}
