package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	examplesTmp := t.TempDir()
	examplesDocTmp := t.TempDir()

	exampleName := "Foo"
	exampleNameLower := strings.ToLower(exampleName)

	err := generate(exampleName, examplesTmp, examplesDocTmp)
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
	assertExampleContent(t, exampleName, filepath.Join(generatedTemplatesDir, exampleNameLower+".go"))
	assertGoModContent(t, exampleName, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, exampleName, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertToolsGoContent(t, exampleName, filepath.Join(generatedTemplatesDir, "tools", "tools.go"))
}

// assert content example file in the docs
func assertExampleDocContent(t *testing.T, exampleName string, exampleDocFile string) {
	content, err := os.ReadFile(exampleDocFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[1], "# "+exampleName)
	assert.Equal(t, data[3], "<!--codeinclude-->")
	assert.Equal(t, data[4], "[Creating a "+exampleName+" container](../../examples/"+lower+"/"+lower+".go)")
	assert.Equal(t, data[5], "<!--/codeinclude-->")
	assert.Equal(t, data[7], "<!--codeinclude-->")
	assert.Equal(t, data[8], "[Test for a "+exampleName+" container](../../examples/"+lower+"/"+lower+"_test.go)")
	assert.Equal(t, data[9], "<!--/codeinclude-->")
}

// assert content example test
func assertExampleTestContent(t *testing.T, exampleName string, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[1], "package "+lower)
	assert.Equal(t, data[8], "func Test"+exampleName+"(t *testing.T) {")
}

// assert content example
func assertExampleContent(t *testing.T, exampleName string, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[1], "package "+lower)
	assert.Equal(t, data[10], "type "+lower+"Container struct {")
	assert.Equal(t, data[14], "func setup"+exampleName+"(ctx context.Context) (*"+lower+"Container, error) {")
	assert.Equal(t, data[26], "\treturn &"+lower+"Container{Container: container}, nil")
}

// assert content go.mod
func assertGoModContent(t *testing.T, exampleName string, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[1], "module github.com/testcontainers/testcontainers-go/examples/"+lower)
}

// assert content Makefile
func assertMakefileContent(t *testing.T, exampleName string, makefile string) {
	content, err := os.ReadFile(makefile)
	assert.Nil(t, err)

	lower := strings.ToLower(exampleName)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[5], "\t$(MAKE) test-"+lower)
}

// assert content tools/tools.go
func assertToolsGoContent(t *testing.T, exampleName string, tools string) {
	content, err := os.ReadFile(tools)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[3], "// This package contains the tool dependencies of the "+exampleName+" example.")
	assert.Equal(t, data[5], "package tools")
}
