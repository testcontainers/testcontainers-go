package workflow

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

type Generator struct{}

// Generate updates github ci workflow
func (g Generator) Generate(ctx context.Context) error {
	examples, modules, err := module.ListExamplesAndModules(ctx)
	if err != nil {
		return fmt.Errorf("list examples and modules: %w", err)
	}

	githubWorkflowsDir := ctx.GithubWorkflowsDir()

	projectDirectories := newProjectDirectories(examples, modules)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(githubWorkflowsDir, "ci.yml")

	return internal_template.GenerateFile(t, exampleFilePath, name, projectDirectories)
}
