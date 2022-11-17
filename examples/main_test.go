package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	rootTmp := t.TempDir()
	examplesTmp := filepath.Join(rootTmp, "examples")
	examplesDocTmp := filepath.Join(rootTmp, "docs", "examples")

	err := os.MkdirAll(examplesTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0777)
	assert.Nil(t, err)
	err = copyInitialConfig(t, rootTmp)
	assert.Nil(t, err)

	example := Example{
		Name:  "foo",
		Image: "docker.io/example/foo:latest",
	}
	exampleNameLower := example.Lower()

	err = generate(example, examplesTmp, examplesDocTmp)
	assert.Nil(t, err)

	templatesDir, err := os.ReadDir(filepath.Join(".", "_template"))
	assert.Nil(t, err)

	exampleDirPath := filepath.Join(examplesTmp, exampleNameLower)
	newExampleDir, err := os.ReadDir(exampleDirPath)
	assert.Nil(t, err)

	exampleDirFileInfo, err := os.Stat(exampleDirPath)
	assert.Nil(t, err) // error nil implies the file exist
	assert.True(t, exampleDirFileInfo.IsDir())

	exampleDocFile := filepath.Join(examplesDocTmp, exampleNameLower+".md")
	_, err = os.Stat(exampleDocFile)
	assert.Nil(t, err) // error nil implies the file exist

	// check the number of template files is equal to examples + 1 (the doc)
	assert.Equal(t, len(newExampleDir)+1, len(templatesDir))

	assertExampleDocContent(t, example, exampleDocFile)

	generatedTemplatesDir := filepath.Join(examplesTmp, exampleNameLower)
	assertExampleTestContent(t, example, filepath.Join(generatedTemplatesDir, exampleNameLower+"_test.go"))
	assertExampleContent(t, example, filepath.Join(generatedTemplatesDir, exampleNameLower+".go"))
	assertGoModContent(t, example, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, example, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertToolsGoContent(t, example, filepath.Join(generatedTemplatesDir, "tools", "tools.go"))
	assertMkdocsExamplesNav(t, example, rootTmp)
}

// assert content example file in the docs
func assertExampleDocContent(t *testing.T, example Example, exampleDocFile string) {
	content, err := os.ReadFile(exampleDocFile)
	assert.Nil(t, err)

	lower := example.Lower()
	title := example.Title()

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "# "+title)
	assert.Equal(t, data[2], "<!--codeinclude-->")
	assert.Equal(t, data[3], "[Creating a "+title+" container](../../examples/"+lower+"/"+lower+".go)")
	assert.Equal(t, data[4], "<!--/codeinclude-->")
	assert.Equal(t, data[6], "<!--codeinclude-->")
	assert.Equal(t, data[7], "[Test for a "+title+" container](../../examples/"+lower+"/"+lower+"_test.go)")
	assert.Equal(t, data[8], "<!--/codeinclude-->")
}

// assert content example test
func assertExampleTestContent(t *testing.T, example Example, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+example.Lower())
	assert.Equal(t, data[7], "func Test"+example.Title()+"(t *testing.T) {")
}

// assert content example
func assertExampleContent(t *testing.T, example Example, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	assert.Nil(t, err)

	lower := example.Lower()
	title := example.Title()
	exampleName := example.Name

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+lower)
	assert.Equal(t, data[8], "// "+lower+"Container represents the "+exampleName+" container type used in the module")
	assert.Equal(t, data[9], "type "+lower+"Container struct {")
	assert.Equal(t, data[13], "// setup"+title+" creates an instance of the "+exampleName+" container type")
	assert.Equal(t, data[14], "func setup"+title+"(ctx context.Context) (*"+lower+"Container, error) {")
	assert.Equal(t, data[16], "\t\tImage: \""+example.Image+"\",")
	assert.Equal(t, data[26], "\treturn &"+lower+"Container{Container: container}, nil")
}

// assert content go.mod
func assertGoModContent(t *testing.T, example Example, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "module github.com/testcontainers/testcontainers-go/examples/"+example.Lower())
}

// assert content Makefile
func assertMakefileContent(t *testing.T, example Example, makefile string) {
	content, err := os.ReadFile(makefile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[4], "\t$(MAKE) test-"+example.Lower())
}

// assert content in the Examples nav from mkdocs.yml
func assertMkdocsExamplesNav(t *testing.T, example Example, rootDir string) {
	config, err := readMkdocsConfig(rootDir)
	assert.Nil(t, err)

	examples := config.Nav[3].Examples
	found := false
	for _, ex := range examples {
		markdownExample := "examples/" + example.Lower() + ".md"
		if markdownExample == ex {
			found = true
		}
	}

	assert.True(t, found)
}

// assert content tools/tools.go
func assertToolsGoContent(t *testing.T, example Example, tools string) {
	content, err := os.ReadFile(tools)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[3], "// This package contains the tool dependencies of the "+example.Name+" example.")
	assert.Equal(t, data[5], "package tools")
}
