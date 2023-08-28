package make

import (
	"path/filepath"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(exampleDir string, exampleName string) error {
	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}
	return internal_template.Generate(t, filepath.Join(exampleDir, "Makefile"), name, exampleName)
}
