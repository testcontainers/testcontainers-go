package workflow

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

// update github ci workflow
func GenerateWorkflow(ctx *context.Context) error {
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
	return Generate(ctx.GithubWorkflowsDir(), examples, modules)
}

func Generate(githubWorkflowsDir string, examples []string, modules []string) error {
	projectDirectories := newProjectDirectories(examples, modules)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(githubWorkflowsDir, "ci.yml")

	return internal_template.GenerateFile(t, exampleFilePath, name, projectDirectories)
}
