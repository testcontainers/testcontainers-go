package mkdocs

import (
	"fmt"
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func GenerateMdFile(filePath string, funcMap template.FuncMap, example any) error {
	name := "module.md.tmpl"
	t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", name, err)
	}

	return internal_template.GenerateFile(t, filePath, name, example)
}
