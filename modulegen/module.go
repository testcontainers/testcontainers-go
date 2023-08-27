package main

import (
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/modfile"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/module"
)

func generateGoModule(ctx *Context, example Example) error {
	exampleDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	err := generateGoFiles(exampleDir, example)
	if err != nil {
		return err
	}
	return generateGoModFile(exampleDir, example)
}

func generateGoFiles(exampleDir string, example Example) error {
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
	}
	return module.GenerateFiles(exampleDir, example.Lower(), funcMap, example)
}

func generateGoModFile(exampleDir string, example Example) error {
	rootCtx, err := getRootContext()
	if err != nil {
		return err
	}
	mkdocsConfig, err := mkdocs.ReadConfig(rootCtx.MkdocsConfigFile())
	if err != nil {
		fmt.Printf(">> could not read MkDocs config: %v\n", err)
		return err
	}
	rootGoModFile := rootCtx.GoModFile()
	directory := "/" + example.ParentDir() + "/" + example.Lower()
	tcVersion := mkdocsConfig.Extra.LatestVersion
	return modfile.GenerateModFile(exampleDir, rootGoModFile, directory, tcVersion)
}
