package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	tests := []struct {
		name                  string
		example               Example
		expectedContainerName string
		expectedEntrypoint    string
		expectedTitle         string
	}{
		{
			name: "Module with title",
			example: Example{
				Name:      "mongoDB",
				IsModule:  true,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedContainerName: "MongoDBContainer",
			expectedEntrypoint:    "RunContainer",
			expectedTitle:         "MongoDB",
		},
		{
			name: "Module without title",
			example: Example{
				Name:     "mongoDB",
				IsModule: true,
				Image:    "mongodb:latest",
			},
			expectedContainerName: "MongodbContainer",
			expectedEntrypoint:    "RunContainer",
			expectedTitle:         "Mongodb",
		},
		{
			name: "Example with title",
			example: Example{
				Name:      "mongoDB",
				IsModule:  false,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedContainerName: "mongoDBContainer",
			expectedEntrypoint:    "runContainer",
			expectedTitle:         "MongoDB",
		},
		{
			name: "Example without title",
			example: Example{
				Name:     "mongoDB",
				IsModule: false,
				Image:    "mongodb:latest",
			},
			expectedContainerName: "mongodbContainer",
			expectedEntrypoint:    "runContainer",
			expectedTitle:         "Mongodb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			example := test.example

			assert.Equal(t, "mongodb", example.Lower())
			assert.Equal(t, test.expectedTitle, example.Title())
			assert.Equal(t, test.expectedContainerName, example.ContainerName())
			assert.Equal(t, test.expectedEntrypoint, example.Entrypoint())
		})
	}
}

func TestExample_Validate(outer *testing.T) {
	outer.Parallel()

	tests := []struct {
		name        string
		example     Example
		expectedErr error
	}{
		{
			name: "only alphabetical characters in name/title",
			example: Example{
				Name:      "AmazingDB",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in name",
			example: Example{
				Name:      "AmazingDB4tw",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in title",
			example: Example{
				Name:      "AmazingDB",
				TitleName: "AmazingDB4tw",
			},
		},
		{
			name: "non-alphanumerical characters in name",
			example: Example{
				Name:      "Amazing DB 4 The Win",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "non-alphanumerical characters in title",
			example: Example{
				Name:      "AmazingDB",
				TitleName: "Amazing DB 4 The Win",
			},
			expectedErr: errors.New("invalid title: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in name",
			example: Example{
				Name:      "1AmazingDB",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in title",
			example: Example{
				Name:      "AmazingDB",
				TitleName: "1AmazingDB",
			},
			expectedErr: errors.New("invalid title: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
	}

	for _, test := range tests {
		outer.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedErr, test.example.Validate())
		})
	}
}

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

func TestGenerateWrongExampleTitle(t *testing.T) {
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
		title string
	}{
		{title: " fooDB"},
		{title: "fooDB "},
		{title: "foo barDB"},
		{title: "foo-barDB"},
		{title: "foo/barDB"},
		{title: "foo\\barDB"},
		{title: "1fooDB"},
		{title: "foo1DB"},
		{title: "-fooDB"},
		{title: "foo-DB"},
	}

	for _, test := range tests {
		example := Example{
			Name:      "foo",
			TitleName: test.title,
			Image:     "docker.io/example/foo:latest",
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
		Name:      "foodb4tw",
		TitleName: "FooDB4TheWin",
		IsModule:  false,
		Image:     "docker.io/example/foodb:latest",
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

func TestGenerateModule(t *testing.T) {
	rootTmp := t.TempDir()
	modulesTmp := filepath.Join(rootTmp, "modules")
	modulesDocTmp := filepath.Join(rootTmp, "docs", "modules")
	githubWorkflowsTmp := filepath.Join(rootTmp, ".github", "workflows")

	err := os.MkdirAll(modulesTmp, 0777)
	assert.Nil(t, err)
	err = os.MkdirAll(modulesDocTmp, 0777)
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
		Name:      "foodb",
		TitleName: "FooDB",
		IsModule:  true,
		Image:     "docker.io/example/foodb:latest",
		TCVersion: "v0.0.0-test",
	}
	exampleNameLower := example.Lower()

	err = generate(example, rootTmp)
	assert.Nil(t, err)

	templatesDir, err := os.ReadDir(filepath.Join(".", "_template"))
	assert.Nil(t, err)

	exampleDirPath := filepath.Join(modulesTmp, exampleNameLower)
	newExampleDir, err := os.ReadDir(exampleDirPath)
	assert.Nil(t, err)

	exampleDirFileInfo, err := os.Stat(exampleDirPath)
	assert.Nil(t, err) // error nil implies the file exist
	assert.True(t, exampleDirFileInfo.IsDir())

	exampleDocFile := filepath.Join(modulesDocTmp, exampleNameLower+".md")
	_, err = os.Stat(exampleDocFile)
	assert.Nil(t, err) // error nil implies the file exist

	exampleWorkflowFile := filepath.Join(githubWorkflowsTmp, "module-"+exampleNameLower+".yml")
	_, err = os.Stat(exampleWorkflowFile)
	assert.Nil(t, err) // error nil implies the file exist

	// check the number of template files is equal to examples + 2 (the doc and the github workflow)
	assert.Equal(t, len(newExampleDir)+2, len(templatesDir))

	assertExampleDocContent(t, example, exampleDocFile)
	assertExampleGithubWorkflowContent(t, example, exampleWorkflowFile)

	generatedTemplatesDir := filepath.Join(modulesTmp, exampleNameLower)
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
		directory := "/" + example.ParentDir() + "/" + example.Lower()
		if directory == ex.Directory {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the main module
	assert.Equal(t, "/", examples[0].Directory, examples)
}

// assert content example file in the docs
func assertExampleDocContent(t *testing.T, example Example, exampleDocFile string) {
	content, err := os.ReadFile(exampleDocFile)
	assert.Nil(t, err)

	lower := example.Lower()
	title := example.Title()

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "# "+title)
	assert.Equal(t, data[2], "## Adding this module to your project dependencies")
	assert.Equal(t, data[4], "Please run the following command to add the "+title+" module to your Go dependencies:")
	assert.Equal(t, data[7], "go get github.com/testcontainers/testcontainers-go/"+example.ParentDir()+"/"+lower)
	assert.Equal(t, data[12], "<!--codeinclude-->")
	assert.Equal(t, data[13], "[Creating a "+title+" container](../../"+example.ParentDir()+"/"+lower+"/"+lower+".go)")
	assert.Equal(t, data[14], "<!--/codeinclude-->")
	assert.Equal(t, data[16], "<!--codeinclude-->")
	assert.Equal(t, data[17], "[Test for a "+title+" container](../../"+example.ParentDir()+"/"+lower+"/"+lower+"_test.go)")
	assert.Equal(t, data[18], "<!--/codeinclude-->")
	assert.Equal(t, data[22], "The "+title+" module exposes one entrypoint function to create the "+title+" container, and this function receives two parameters:")
	assert.True(t, strings.HasSuffix(data[25], "(*"+title+"Container, error)"))
	assert.Equal(t, "for "+title+". E.g. `testcontainers.WithImage(\""+example.Image+"\")`.", data[38])
}

// assert content example test
func assertExampleTestContent(t *testing.T, example Example, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+example.Lower())
	assert.Equal(t, data[7], "func Test"+example.Title()+"(t *testing.T) {")
	assert.Equal(t, data[10], "\tcontainer, err := setup"+example.Title()+"(ctx)")
}

// assert content example
func assertExampleContent(t *testing.T, example Example, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	assert.Nil(t, err)

	lower := example.Lower()
	containerName := example.ContainerName()
	exampleName := example.Title()
	entrypoint := example.Entrypoint()

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[0], "package "+lower)
	assert.Equal(t, data[8], "// "+containerName+" represents the "+exampleName+" container type used in the module")
	assert.Equal(t, data[9], "type "+containerName+" struct {")
	assert.Equal(t, data[13], "// "+entrypoint+" creates an instance of the "+exampleName+" container type")
	assert.Equal(t, data[14], "func "+entrypoint+"(ctx context.Context, opts ...testcontainers.CustomizeRequestOption) (*"+containerName+", error) {")
	assert.Equal(t, data[16], "\t\tImage: \""+example.Image+"\",")
	assert.Equal(t, data[31], "\treturn &"+containerName+"{Container: container}, nil")
}

// assert content GitHub workflow for the example
func assertExampleGithubWorkflowContent(t *testing.T, example Example, exampleWorkflowFile string) {
	content, err := os.ReadFile(exampleWorkflowFile)
	assert.Nil(t, err)

	lower := example.Lower()
	title := example.Title()

	data := strings.Split(string(content), "\n")
	assert.Equal(t, "name: "+title+" "+example.Type()+" pipeline", data[0])
	assert.Equal(t, "  test-"+lower+":", data[19])
	assert.Equal(t, "          go-version: ${{ matrix.go-version }}", data[29])
	assert.Equal(t, "        working-directory: ./"+example.ParentDir()+"/"+lower, data[36])
	assert.Equal(t, "        working-directory: ./"+example.ParentDir()+"/"+lower, data[40])
	assert.Equal(t, "        working-directory: ./"+example.ParentDir()+"/"+lower, data[44])
	assert.Equal(t, "          paths: \"**/TEST-"+lower+"*.xml\"", data[54])
}

// assert content go.mod
func assertGoModContent(t *testing.T, example Example, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, "module github.com/testcontainers/testcontainers-go/"+example.ParentDir()+"/"+example.Lower(), data[0])
	assert.Equal(t, "\tgithub.com/testcontainers/testcontainers-go "+example.TCVersion, data[5])
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

	parentDir := example.ParentDir()

	examples := config.Nav[4].Examples
	expectedEntries := originalConfig.Nav[4].Examples
	if example.IsModule {
		examples = config.Nav[3].Modules
		expectedEntries = originalConfig.Nav[3].Modules
	}

	assert.Equal(t, len(expectedEntries)+1, len(examples))

	// the example should be in the nav
	found := false
	for _, ex := range examples {
		markdownExample := example.ParentDir() + "/" + example.Lower() + ".md"
		if markdownExample == ex {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the index
	assert.Equal(t, parentDir+"/index.md", examples[0], examples)
}

// assert content tools/tools.go
func assertToolsGoContent(t *testing.T, example Example, tools string) {
	content, err := os.ReadFile(tools)
	assert.Nil(t, err)

	data := strings.Split(string(content), "\n")
	assert.Equal(t, data[3], "// This package contains the tool dependencies of the "+example.Title()+" "+example.Type()+".")
	assert.Equal(t, data[5], "package tools")
}
