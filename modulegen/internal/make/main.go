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

	exampleFilePath := filepath.Join(exampleDir, "Makefile")

	return internal_template.GenerateFile(t, exampleFilePath, name, exampleName)
}
