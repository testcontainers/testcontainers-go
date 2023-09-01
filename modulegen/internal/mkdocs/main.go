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
		"Entrypoint":    func() string { return tcModule.Entrypoint() },
		"ContainerName": func() string { return tcModule.ContainerName() },
		"ParentDir":     func() string { return tcModule.ParentDir() },
		"ToLower":       func() string { return tcModule.Lower() },
		"Title":         func() string { return tcModule.Title() },
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

func CopyConfig(configFile string, tmpFile string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
