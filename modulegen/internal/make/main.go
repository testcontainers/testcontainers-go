package make

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

// Generator is a struct that contains the logic to generate the Makefile.
type Generator struct{}

// AddModule update Makefile with the new module
func (g Generator) AddModule(ctx context.Context, tcModule context.TestcontainersModule) error {
	moduleDir := filepath.Join(ctx.RootDir, tcModule.ParentDir(), tcModule.Lower())
	moduleName := tcModule.Lower()

	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	moduleFilePath := filepath.Join(moduleDir, "Makefile")

	return internal_template.GenerateFile(t, moduleFilePath, name, moduleName)
}

// GenerateMakefile creates Makefile for example
func GenerateMakefile(ctx context.Context, tcModule context.TestcontainersModule) error {
	moduleDir := filepath.Join(ctx.RootDir, tcModule.ParentDir(), tcModule.Lower())
	moduleName := tcModule.Lower()

	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	moduleFilePath := filepath.Join(moduleDir, "Makefile")

	return internal_template.GenerateFile(t, moduleFilePath, name, moduleName)
}
