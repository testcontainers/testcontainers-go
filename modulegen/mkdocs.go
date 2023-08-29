package main

import (
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
)

// update examples in mkdocs
func generateMkdocs(ctx *Context, example Example) error {
	exampleMdFile := filepath.Join(ctx.DocsDir(), example.ParentDir(), example.Lower()+".md")
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
	}
	err := mkdocs.GenerateMdFile(exampleMdFile, funcMap, example)
	if err != nil {
		return err
	}
	exampleMd := example.ParentDir() + "/" + example.Lower() + ".md"
	indexMd := example.ParentDir() + "/index.md"
	return mkdocs.UpdateConfig(ctx.MkdocsConfigFile(), example.IsModule, exampleMd, indexMd)
}
