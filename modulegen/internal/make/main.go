package make

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

// creates Makefile for example
func GenerateMakefile(ctx *context.Context, example context.Example) error {
	exampleDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	exampleName := example.Lower()
	return Generate(exampleDir, exampleName)
}

func Generate(exampleDir string, exampleName string) error {
	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(exampleDir, "Makefile")

	return internal_template.GenerateFile(t, exampleFilePath, name, exampleName)
}
