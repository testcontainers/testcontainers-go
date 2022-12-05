package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateWrongExampleName(t *testing.T) {
	rootTmp := t.TempDir()
	examplesTmp := filepath.Join(rootTmp, "examples")
	examplesDocTmp := filepath.Join(rootTmp, "docs", "examples")
	githubWorkflowsTmp := filepath.Join(rootTmp, ".github", "workflows")

	err := os.MkdirAll(examplesTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, rootTmp)
	assert.Nil(t, err)

	tests := []struct {
		name string
	}{
		{name: " foo"},
		{name: "foo "},
		{name: "foo bar"},
		{name: "foo-bar"},
		{name: "foo/bar"},
		{name: "foo\\bar"},
		{name: "1foo"},
		{name: "foo1"},
		{name: "-foo"},
		{name: "foo-"},
	}

	for _, test := range tests {
		example := Example{
			Name:      test.name,
			Image:     "docker.io/example/" + test.name + ":latest",
			TCVersion: "v0.0.0-test",
		}

		err = generate(example, rootTmp)
		assert.Error(t, err)
	}
}

func TestGenerate(t *testing.T) {
	rootTmp := t.TempDir()
	examplesTmp := filepath.Join(rootTmp, "examples")
	examplesDocTmp := filepath.Join(rootTmp, "docs", "examples")
	githubWorkflowsTmp := filepath.Join(rootTmp, ".github", "workflows")

	err := os.MkdirAll(examplesTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, rootTmp)
	assert.Nil(t, err)

	originalConfig, err := readMkdocsConfig(rootTmp)
	assert.Nil(t, err)

	err = copyInitialDependabotConfig(t, rootTmp)
	assert.Nil(t, err)

	originalDependabotConfig, err := readDependabotConfig(rootTmp)
	assert.Nil(t, err)

	example := Example{
		Name:      "foo",
		Image:     "docker.io/example/foo:latest",
		TCVersion: "v0.0.0-test",
	}
	exampleNameLower := example.Lower()

	err = generate(example, rootTmp)
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

	exampleWorkflowFile := filepath.Join(githubWorkflowsTmp, exampleNameLower+"-example.yml")
	_, err = os.Stat(exampleWorkflowFile)
	assert.Nil(t, err) // error nil implies the file exist

	// check the number of template files is equal to examples + 2 (the doc and the github workflow)
	assert.Equal(t, len(newExampleDir)+2, len(templatesDir))

	assertExampleDocContent(t, example, exampleDocFile)
	assertExampleGithubWorkflowContent(t, example, exampleWorkflowFile)

	generatedTemplatesDir := filepath.Join(examplesTmp, exampleNameLower)
	assertExampleTestContent(t, example, filepath.Join(generatedTemplatesDir, exampleNameLower+"_test.go"))
	assertExampleContent(t, example, filepath.Join(generatedTemplatesDir, exampleNameLower+".go"))
	assertGoModContent(t, example, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, example, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertToolsGoContent(t, example, filepath.Join(generatedTemplatesDir, "tools", "tools.go"))
	assertMkdocsExamplesNav(t, example, originalConfig, rootTmp)
	assertDependabotExamplesUpdates(t, example, originalDependabotConfig, rootTmp)
}

// assert content in the Examples nav from mkdocs.yml
func assertDependabotExamplesUpdates(t *testing.T, example Example, originalConfig *DependabotConfig, rootDir string) {
	config, err := readDependabotConfig(rootDir)
	assert.Nil(t, err)

	examples := config.Updates

	assert.Equal(t, len(originalConfig.Updates)+1, len(examples))

	// the example should be in the dependabot updates
	found := false
	for _, ex := range examples {
		directory := "/examples/" + example.Lower()
		if directory == ex.Directory {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the main module
	assert.Equal(t, "/", examples[0].Directory, examples)
	// second item is the e2e module
	assert.Equal(t, "/e2e", examples[1].Directory, examples)
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

// assert content GitHub workflow for the example
func assertExampleGithubWorkflowContent(t *testing.T, example Example, exampleWorkflowFile string) {
	content, err := os.ReadFile(exampleWorkflowFile)
	assert.Nil(t, err)

	lower := example.Lower()
	title := example.Title()

	data := strings.Split(string(content), "\n")
	assert.Equal(t, "name: "+title+" example pipeline", data[0])
	assert.Equal(t, "  test-"+lower+":", data[5])
	assert.Equal(t, "          go-version: ${{ matrix.go-version }}", data[15])
	assert.Equal(t, "        working-directory: ./examples/"+lower, data[22])
	assert.Equal(t, "        working-directory: ./examples/"+lower, data[26])
	assert.Equal(t, "        working-directory: ./examples/"+lower, data[30])
	assert.Equal(t, "          paths: \"**/TEST-"+lower+"*.xml\"", data[40])
}

// assert content go.mod
func assertGoModContent(t *testing.T, example Example, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, "module github.com/testcontainers/testcontainers-go/examples/"+example.Lower(), data[0])
	assert.Equal(t, "\tgithub.com/testcontainers/testcontainers-go v"+example.TCVersion, data[5])
}

// assert content Makefile
func assertMakefileContent(t *testing.T, example Example, makefile string) {
	content, err := os.ReadFile(makefile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[4], "\t$(MAKE) test-"+example.Lower())
}

// assert content in the Examples nav from mkdocs.yml
func assertMkdocsExamplesNav(t *testing.T, example Example, originalConfig *MkDocsConfig, rootDir string) {
	config, err := readMkdocsConfig(rootDir)
	assert.Nil(t, err)

	examples := config.Nav[3].Examples

	assert.Equal(t, len(originalConfig.Nav[3].Examples)+1, len(examples))

	// the example should be in the nav
	found := false
	for _, ex := range examples {
		markdownExample := "examples/" + example.Lower() + ".md"
		if markdownExample == ex {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the index
	assert.Equal(t, "examples/index.md", examples[0], examples)
}

// assert content tools/tools.go
func assertToolsGoContent(t *testing.T, example Example, tools string) {
	content, err := os.ReadFile(tools)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[3], "// This package contains the tool dependencies of the "+example.Name+" example.")
	assert.Equal(t, data[5], "package tools")
}
