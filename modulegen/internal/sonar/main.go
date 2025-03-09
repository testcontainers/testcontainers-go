package sonar

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

// Generator is a struct that contains the logic to generate the sonar-project.properties file.
type Generator struct{}

// Generate updates sonar-project.properties
func (g Generator) Generate(ctx context.Context, examples []string, modules []string) error {
	mkdocsConfig, err := mkdocs.ReadConfig(ctx.MkdocsConfigFile())
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	tcVersion := mkdocsConfig.Extra.LatestVersion
	config := newConfig(tcVersion, examples, modules)
	name := "sonar-project.properties.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return fmt.Errorf("parse files: %w", err)
	}

	return internal_template.GenerateFile(t, ctx.SonarProjectFile(), name, config)
}
