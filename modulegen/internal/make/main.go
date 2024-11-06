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
		return fmt.Errorf("parse makefile template: %w", err)
	}

	moduleFilePath := filepath.Join(moduleDir, "Makefile")
	if err := internal_template.GenerateFile(t, moduleFilePath, name, moduleName); err != nil {
		return fmt.Errorf("generate file: %w", err)
	}

	return nil
}
