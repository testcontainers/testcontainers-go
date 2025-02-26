package sonar

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

type Generator struct{}

// Generate updates sonar-project.properties
func (g Generator) Generate(ctx context.Context) error {
	examples, err := ctx.GetExamples()
	if err != nil {
		return err
	}
	modules, err := ctx.GetModules()
	if err != nil {
		return err
	}
	mkdocsConfig, err := mkdocs.ReadConfig(ctx.MkdocsConfigFile())
	if err != nil {
		fmt.Printf(">> could not read MkDocs config: %v\n", err)
		return err
	}
	tcVersion := mkdocsConfig.Extra.LatestVersion
	config := newConfig(tcVersion, examples, modules)
	name := "sonar-project.properties.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	return internal_template.GenerateFile(t, ctx.SonarProjectFile(), name, config)
}

// Refresh refresh the sonar-project.properties file
func (g Generator) Refresh(ctx context.Context, tcModules []context.TestcontainersModule) error {
	return g.Generate(ctx)
}
