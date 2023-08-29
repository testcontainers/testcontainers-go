package workflow

import (
	"os"
	"path/filepath"
	"text/template"

	internal "github.com/testcontainers/testcontainers-go/modulegen/internal"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func Generate(githubWorkflowsDir string, examples []string, modules []string) error {
	projectDirectories := internal.NewProjectDirectories(examples, modules)
	name := "ci.yml.tmpl"
	t, err := template.New(name).ParseFiles(filepath.Join("_template", name))
	if err != nil {
		return err
	}

	exampleFilePath := filepath.Join(githubWorkflowsDir, "ci.yml")

	err = os.MkdirAll(filepath.Dir(exampleFilePath), 0o755)
	if err != nil {
		return err
	}

	exampleFile, err := os.Create(exampleFilePath)
	if err != nil {
		return err
	}
	defer exampleFile.Close()

	return internal_template.Generate(t, exampleFile, name, projectDirectories)
}
