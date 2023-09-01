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

type Generator struct{}

// AddModule creates the go.mod file for the module
func (g Generator) AddModule(ctx context.Context, tcModule context.TestcontainersModule) error {
	moduleDir := filepath.Join(ctx.RootDir, tcModule.ParentDir(), tcModule.Lower())
	err := generateGoFiles(moduleDir, tcModule)
	if err != nil {
		return err
	}
	return generateGoModFile(moduleDir, tcModule)
}

func generateGoFiles(moduleDir string, tcModule context.TestcontainersModule) error {
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return tcModule.Entrypoint() },
		"ContainerName": func() string { return tcModule.ContainerName() },
		"ParentDir":     func() string { return tcModule.ParentDir() },
		"ToLower":       func() string { return tcModule.Lower() },
		"Title":         func() string { return tcModule.Title() },
	}
	return GenerateFiles(moduleDir, tcModule.Lower(), funcMap, tcModule)
}

func generateGoModFile(moduleDir string, tcModule context.TestcontainersModule) error {
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
	directory := "/" + tcModule.ParentDir() + "/" + tcModule.Lower()
	tcVersion := mkdocsConfig.Extra.LatestVersion
	return modfile.GenerateModFile(moduleDir, rootGoModFile, directory, tcVersion)
}

func GenerateFiles(moduleDir string, moduleName string, funcMap template.FuncMap, tcModule any) error {
	for _, tmpl := range []string{"example_test.go", "example.go"} {
		name := tmpl + ".tmpl"
		t, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join("_template", name))
		if err != nil {
			return err
		}
		moduleFilePath := filepath.Join(moduleDir, strings.ReplaceAll(tmpl, "example", moduleName))

		err = internal_template.GenerateFile(t, moduleFilePath, name, tcModule)
		if err != nil {
			return err
		}
	}
	return nil
}
