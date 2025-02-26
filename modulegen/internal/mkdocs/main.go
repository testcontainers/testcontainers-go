package mkdocs

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

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
		return err
	}
	moduleMd := tcModule.ParentDir() + "/" + tcModule.Lower() + ".md"
	indexMd := tcModule.ParentDir() + "/index.md"

	configFile := ctx.MkdocsConfigFile()
	isModule := tcModule.IsModule

	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	config.addModule(isModule, moduleMd, indexMd)
	return writeConfig(configFile, config)
}

// Refresh refresh the mkdocs config file for all the modules,
// excluding compose as it has its own page in the docs.
func (g Generator) Generate(ctx context.Context, examples []string, modules []string) error {
	configFile := ctx.MkdocsConfigFile()
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}

	for _, module := range modules {
		if module == "compose" {
			continue
		}

		moduleMd := "modules/" + strings.ToLower(module) + ".md"
		indexMd := "modules/index.md"

		config.addModule(true, moduleMd, indexMd)
	}

	for _, example := range examples {
		exampleMd := "examples/" + strings.ToLower(example) + ".md"
		indexMd := "examples/index.md"

		config.addModule(false, exampleMd, indexMd)
	}

	return writeConfig(configFile, config)
}

func CopyConfig(configFile string, tmpFile string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
