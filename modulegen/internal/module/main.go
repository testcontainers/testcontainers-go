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
func (g Generator) AddModule(ctx context.Context, m context.TestcontainersModule) error {
	moduleDir := filepath.Join(ctx.RootDir, m.ParentDir(), m.Lower())
	err := generateGoFiles(moduleDir, m)
	if err != nil {
		return err
	}
	return generateGoModFile(moduleDir, m)
}

func generateGoFiles(moduleDir string, m context.TestcontainersModule) error {
	funcMap := template.FuncMap{
		"Entrypoint":    func() string { return m.Entrypoint() },
		"ContainerName": func() string { return m.ContainerName() },
		"ParentDir":     func() string { return m.ParentDir() },
		"ToLower":       func() string { return m.Lower() },
		"Title":         func() string { return m.Title() },
	}
	return GenerateFiles(moduleDir, m.Lower(), funcMap, m)
}

func generateGoModFile(moduleDir string, m context.TestcontainersModule) error {
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
	directory := "/" + m.ParentDir() + "/" + m.Lower()
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
