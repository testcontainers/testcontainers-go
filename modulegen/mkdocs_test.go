package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
)

func TestGetMkDocsConfigFile(t *testing.T) {
	tmpCtx := context.New(filepath.Join(t.TempDir(), "testcontainers-go"))
	cfgFile := tmpCtx.MkdocsConfigFile()
	err := os.MkdirAll(tmpCtx.RootDir, 0o777)
	assert.NilError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0o777)
	assert.NilError(t, err)

	file := tmpCtx.MkdocsConfigFile()
	assert.Assert(t, file != "")

	assert.Check(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", "mkdocs.yml")))
}

func TestReadMkDocsConfig(t *testing.T) {
	tmpCtx := context.New(filepath.Join(t.TempDir(), "testcontainers-go"))
	err := os.MkdirAll(tmpCtx.RootDir, 0o777)
	assert.NilError(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	assert.NilError(t, err)

	config, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	assert.NilError(t, err)
	assert.Assert(t, config != nil)

	assert.Check(t, is.Equal("Testcontainers for Go", config.SiteName))
	assert.Check(t, is.Equal("https://github.com/testcontainers/testcontainers-go", config.RepoURL))
	assert.Check(t, is.Equal("edit/main/docs/", config.EditURI))

	// theme
	theme := config.Theme
	assert.Check(t, is.Equal("material", theme.Name))

	// nav bar
	nav := config.Nav
	assert.Check(t, is.Equal("index.md", nav[0].Home))
	assert.Check(t, len(nav[2].Features) != 0)
	assert.Check(t, len(nav[3].Modules) != 0)
	assert.Check(t, len(nav[4].Examples) != 0)
}

func TestNavItems(t *testing.T) {
	ctx := getTestRootContext(t)
	examples, err := ctx.GetExamples()
	assert.NilError(t, err)
	examplesDocs, err := ctx.GetExamplesDocs()
	assert.NilError(t, err)

	// we have to remove the index.md file from the examples docs
	assert.Check(t, is.Len(examples, len(examplesDocs)-1))

	// all example modules exist in the documentation
	for _, example := range examples {
		found := false
		for _, exampleDoc := range examplesDocs {
			markdownName := example + ".md"

			if markdownName == exampleDoc {
				found = true
				continue
			}
		}
		assert.Check(t, found, "example %s is not present in the docs", example)
	}
}

func copyInitialMkdocsConfig(t *testing.T, tmpCtx context.Context) error {
	ctx := getTestRootContext(t)
	return mkdocs.CopyConfig(ctx.MkdocsConfigFile(), tmpCtx.MkdocsConfigFile())
}
