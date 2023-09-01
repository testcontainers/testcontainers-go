package workflow

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

type Generator struct{}

// Generate updates github ci workflow
func (g Generator) Generate(ctx context.Context) error {
	rootCtx, err := context.GetRootContext()
	if err != nil {
		return err
	}
	examples, err := rootCtx.GetExamples()
	if err != nil {
		return err
	}
	modules, err := rootCtx.GetModules()
	if err != nil {
		return err
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
