package make

import (
	"os"
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

	err = os.MkdirAll(filepath.Dir(exampleFilePath), 0o755)
	if err != nil {
		return err
	}

	exampleFile, err := os.Create(exampleFilePath)
	if err != nil {
		return err
	}
	defer exampleFile.Close()

	return internal_template.Generate(t, exampleFile, name, exampleName)
}
