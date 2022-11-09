package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TestGenerate(t *testing.T) {
	examplesTmp := t.TempDir()
	examplesDocTmp := t.TempDir()

	exampleName := "foo"
	exampleImage := "docker.io/example/foo:latest"
	exampleNameLower := strings.ToLower(exampleName)

	err := generate(exampleName, exampleImage, examplesTmp, examplesDocTmp)
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

	assertExampleDocContent(t, exampleName, exampleDocFile)

	generatedTemplatesDir := filepath.Join(examplesTmp, exampleNameLower)
	assertExampleTestContent(t, exampleName, filepath.Join(generatedTemplatesDir, exampleNameLower+"_test.go"))
	assertExampleContent(t, exampleName, exampleImage, filepath.Join(generatedTemplatesDir, exampleNameLower+".go"))
	assertGoModContent(t, exampleName, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, exampleName, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertToolsGoContent(t, exampleName, filepath.Join(generatedTemplatesDir, "tools", "tools.go"))
}

// assert content example file in the docs
func assertExampleDocContent(t *testing.T, exampleName string, exampleDocFile string) {
	content, err := os.ReadFile(exampleDocFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)
	title := cases.Title(language.Und, cases.NoLower).String(exampleName)

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
func assertExampleTestContent(t *testing.T, exampleName string, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)
	title := cases.Title(language.Und, cases.NoLower).String(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+lower)
	assert.Equal(t, data[7], "func Test"+title+"(t *testing.T) {")
}

// assert content example
func assertExampleContent(t *testing.T, exampleName string, exampleImage string, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)
	title := cases.Title(language.Und, cases.NoLower).String(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+lower)
	assert.Equal(t, data[8], "// "+lower+"Container represents the "+exampleName+" container type used in the module")
	assert.Equal(t, data[9], "type "+lower+"Container struct {")
	assert.Equal(t, data[13], "// setup"+title+" creates an instance of the "+exampleName+" container type")
	assert.Equal(t, data[14], "func setup"+title+"(ctx context.Context) (*"+lower+"Container, error) {")
	assert.Equal(t, data[16], "\t\tImage: \""+exampleImage+"\",")
	assert.Equal(t, data[26], "\treturn &"+lower+"Container{Container: container}, nil")
}

// assert content go.mod
func assertGoModContent(t *testing.T, exampleName string, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "module github.com/testcontainers/testcontainers-go/examples/"+lower)
}

// assert content Makefile
func assertMakefileContent(t *testing.T, exampleName string, makefile string) {
	content, err := os.ReadFile(makefile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[4], "\t$(MAKE) test-"+lower)
}

// assert content tools/tools.go
func assertToolsGoContent(t *testing.T, exampleName string, tools string) {
	content, err := os.ReadFile(tools)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[3], "// This package contains the tool dependencies of the "+exampleName+" example.")
	assert.Equal(t, data[5], "package tools")
}
