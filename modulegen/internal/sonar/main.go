package sonar

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

type Generator struct{}

// Generate updates sonar-project.properties
func (g Generator) Generate(ctx context.Context) error {
	examples, modules, err := module.ListExamplesAndModules(ctx)
	if err != nil {
		return fmt.Errorf("list examples and modules: %w", err)
	}

	rootCtx, err := context.GetRootContext()
	if err != nil {
		return err
	}

	mkdocsConfig, err := mkdocs.ReadConfig(rootCtx.MkdocsConfigFile())
	if err != nil {
		return fmt.Errorf("read MkDocs config: %w", err)
	}
	tcVersion := mkdocsConfig.Extra.LatestVersion
	config := newConfig(tcVersion, examples, modules)
	name := "sonar-project.properties.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return fmt.Errorf("parse %s template: %w", name, err)
	}

	return internal_template.GenerateFile(t, ctx.SonarProjectFile(), name, config)
}
