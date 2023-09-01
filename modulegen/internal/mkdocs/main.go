package mkdocs

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

type Generator struct{}

// AddModule update modules in mkdocs
func (g Generator) AddModule(ctx *context.Context, m context.TestcontainersModule) error {
	moduleMdFile := filepath.Join(ctx.DocsDir(), m.ParentDir(), m.Lower()+".md")
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return m.Entrypoint() },
		"ContainerName": func() string { return m.ContainerName() },
		"ParentDir":     func() string { return m.ParentDir() },
		"ToLower":       func() string { return m.Lower() },
		"Title":         func() string { return m.Title() },
	}
	err := GenerateMdFile(moduleMdFile, funcMap, m)
	if err != nil {
		return err
	}
	moduleMd := m.ParentDir() + "/" + m.Lower() + ".md"
	indexMd := m.ParentDir() + "/index.md"

	configFile := ctx.MkdocsConfigFile()
	isModule := m.IsModule

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
