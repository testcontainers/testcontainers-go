package module

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/modfile"
	internal_template "github.com/testcontainers/testcontainers-go/modulegen/internal/template"
)

func GenerateGoModule(ctx *context.Context, example context.Example) error {
	exampleDir := filepath.Join(ctx.RootDir, example.ParentDir(), example.Lower())
	err := generateGoFiles(exampleDir, example)
	if err != nil {
		return err
	}
	return generateGoModFile(exampleDir, example)
}

func generateGoFiles(exampleDir string, example context.Example) error {
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return example.Entrypoint() },
		"ContainerName": func() string { return example.ContainerName() },
		"ParentDir":     func() string { return example.ParentDir() },
		"ToLower":       func() string { return example.Lower() },
		"Title":         func() string { return example.Title() },
	}
	return GenerateFiles(exampleDir, example.Lower(), funcMap, example)
}

func generateGoModFile(exampleDir string, example context.Example) error {
	rootCtx, err := context.GetRootContext()
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

func GenerateFiles(exampleDir string, exampleName string, funcMap template.FuncMap, example any) error {
	for _, tmpl := range []string{"example_test.go", "example.go"} {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}
		exampleFilePath := filepath.Join(exampleDir, strings.ReplaceAll(tmpl, "example", exampleName))

		err = internal_template.GenerateFile(t, exampleFilePath, name, example)
		if err != nil {
			return err
		}
	}
	return nil
}
