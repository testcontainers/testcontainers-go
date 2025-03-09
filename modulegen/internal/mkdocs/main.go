package mkdocs

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// Generator is a struct that contains the logic to generate the mkdocs config file.
type Generator struct{}

// AddModule update modules in mkdocs
func (g Generator) AddModule(ctx context.Context, tcModule context.TestcontainersModule) error {
	moduleMdFile := filepath.Join(ctx.DocsDir(), tcModule.ParentDir(), tcModule.Lower()+".md")
	funcMap := template.FuncMap{
		"Entrypoint":    tcModule.Entrypoint,
		"ContainerName": tcModule.ContainerName,
		"ParentDir":     tcModule.ParentDir,
		"ToLower":       tcModule.Lower,
		"Title":         tcModule.Title,
	}
	err := GenerateMdFile(moduleMdFile, funcMap, tcModule)
	if err != nil {
		return fmt.Errorf("generate md file: %w", err)
	}
	moduleMd := filepath.Join(tcModule.ParentDir(), tcModule.Lower()+".md")
	indexMd := filepath.Join(tcModule.ParentDir(), "index.md")

	configFile := ctx.MkdocsConfigFile()
	isModule := tcModule.IsModule

	config, err := ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	config.addModule(isModule, moduleMd, indexMd)
	return writeConfig(configFile, config)
}

// Generate refresh the mkdocs config file for all the modules,
// excluding compose as it has its own page in the docs.
func (g Generator) Generate(ctx context.Context, examples []string, modules []string) error {
	configFile := ctx.MkdocsConfigFile()
	config, err := ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	for _, module := range modules {
		// The compose module has its own page in the docs.
		if module == "compose" {
			continue
		}

		moduleMd := filepath.Join("modules", strings.ToLower(module)+".md")
		indexMd := filepath.Join("modules", "index.md")

		config.addModule(true, moduleMd, indexMd)
	}

	for _, example := range examples {
		exampleMd := filepath.Join("examples", strings.ToLower(example)+".md")
		indexMd := filepath.Join("examples", "index.md")

		config.addModule(false, exampleMd, indexMd)
	}

	return writeConfig(configFile, config)
}

// CopyConfig helper function to copy the mkdocs config file to a another file
// in the tests.
func CopyConfig(configFile string, tmpFile string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	return writeConfig(tmpFile, config)
}
