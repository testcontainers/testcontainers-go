package mkdocs

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
)

// update examples in mkdocs
func GenerateMkdocs(ctx *context.Context, example context.Example) error {
	exampleMdFile := filepath.Join(ctx.DocsDir(), example.ParentDir(), example.Lower()+".md")
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
	}
	err := GenerateMdFile(exampleMdFile, funcMap, example)
	if err != nil {
		return err
	}
	exampleMd := example.ParentDir() + "/" + example.Lower() + ".md"
	indexMd := example.ParentDir() + "/index.md"
	return UpdateConfig(ctx.MkdocsConfigFile(), example.IsModule, exampleMd, indexMd)
}

func UpdateConfig(configFile string, isModule bool, exampleMd string, indexMd string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	config.addExample(isModule, exampleMd, indexMd)
	return writeConfig(configFile, config)
}

func CopyConfig(configFile string, tmpFile string) error {
	config, err := ReadConfig(configFile)
	if err != nil {
		return err
	}
	return writeConfig(tmpFile, config)
}
