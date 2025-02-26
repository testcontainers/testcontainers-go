package make

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

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

// Refresh refresh the Makefile in all the modules
func (g Generator) Refresh(ctx context.Context, tcModules []context.TestcontainersModule) error {
	for _, tcModule := range tcModules {
		err := g.AddModule(ctx, tcModule)
		if err != nil {
			return fmt.Errorf("add module %s: %w", tcModule.Name, err)
		}
	}
	return nil
}

// creates Makefile for example
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
