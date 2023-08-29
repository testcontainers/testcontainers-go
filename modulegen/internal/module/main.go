package module

import (
	"path/filepath"
	"strings"
	"text/template"

	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func GenerateFiles(exampleDir string, exampleName string, funcMap template.FuncMap, example any) error {
	for _, tmpl := range []string{"example_test.go", "example.go"} {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}
		exampleFilePath := filepath.Join(exampleDir, strings.ReplaceAll(tmpl, "example", exampleName))

		err = internal_template.Generate(t, exampleFilePath, name, example)
		if err != nil {
			return err
		}
	}
	return nil
}
