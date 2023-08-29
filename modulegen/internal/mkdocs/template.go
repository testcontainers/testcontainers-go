package mkdocs

import (
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func GenerateMdFile(filePath string, funcMap template.FuncMap, example any) error {
	name := "example.md.tmpl"
	t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}
	return internal_template.Generate(t, filePath, name, example)
}
