package make

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

// creates Makefile for example
func GenerateMakefile(ctx *context.Context, m context.TestcontainersModule) error {
	moduleDir := filepath.Join(ctx.RootDir, m.ParentDir(), m.Lower())
	moduleName := m.Lower()
	return Generate(moduleDir, moduleName)
}

func Generate(moduleDir string, moduleName string) error {
	name := "Makefile.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	moduleFilePath := filepath.Join(moduleDir, "Makefile")

	return internal_template.GenerateFile(t, moduleFilePath, name, moduleName)
}
