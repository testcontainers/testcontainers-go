package mkdocs

import (
	"path/filepath"
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
func (g Generator) Refresh(ctx context.Context, tcModules []context.TestcontainersModule) error {
	configFile := ctx.MkdocsConfigFile()
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}

	for _, tcModule := range tcModules {
		if tcModule.Name == "compose" {
			continue
		}

		isModule := tcModule.IsModule
		moduleMd := tcModule.ParentDir() + "/" + tcModule.Lower() + ".md"
		indexMd := tcModule.ParentDir() + "/index.md"

		config.addModule(isModule, moduleMd, indexMd)
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
