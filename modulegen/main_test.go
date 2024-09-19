package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
)

func TestModule(t *testing.T) {
	tests := []struct {
		name                  string
		module                context.TestcontainersModule
		expectedContainerName string
		expectedEntrypoint    string
		expectedTitle         string
	}{
		{
			name: "Module with title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  true,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedContainerName: "MongoDBContainer",
			expectedEntrypoint:    "Run",
			expectedTitle:         "MongoDB",
		},
		{
			name: "Module without title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: true,
				Image:    "mongodb:latest",
			},
			expectedContainerName: "MongodbContainer",
			expectedEntrypoint:    "Run",
			expectedTitle:         "Mongodb",
		},
		{
			name: "Example with title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  false,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedContainerName: "mongoDBContainer",
			expectedEntrypoint:    "run",
			expectedTitle:         "MongoDB",
		},
		{
			name: "Example without title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: false,
				Image:    "mongodb:latest",
			},
			expectedContainerName: "mongodbContainer",
			expectedEntrypoint:    "run",
			expectedTitle:         "Mongodb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := test.module

			assert.Equal(t, "mongodb", module.Lower())
			assert.Equal(t, test.expectedTitle, module.Title())
			assert.Equal(t, test.expectedContainerName, module.ContainerName())
			assert.Equal(t, test.expectedEntrypoint, module.Entrypoint())
		})
	}
}

func TestModule_Validate(outer *testing.T) {
	outer.Parallel()

	tests := []struct {
		name        string
		module      context.TestcontainersModule
		expectedErr error
	}{
		{
			name: "only alphabetical characters in name/title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in name",
			module: context.TestcontainersModule{
				Name:      "AmazingDB4tw",
				TitleName: "AmazingDB",
			},
		},
		{
			name: "alphanumerical characters in title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "AmazingDB4tw",
			},
		},
		{
			name: "non-alphanumerical characters in name",
			module: context.TestcontainersModule{
				Name:      "Amazing DB 4 The Win",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "non-alphanumerical characters in title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "Amazing DB 4 The Win",
			},
			expectedErr: errors.New("invalid title: Amazing DB 4 The Win. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in name",
			module: context.TestcontainersModule{
				Name:      "1AmazingDB",
				TitleName: "AmazingDB",
			},
			expectedErr: errors.New("invalid name: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
		{
			name: "leading numerical character in title",
			module: context.TestcontainersModule{
				Name:      "AmazingDB",
				TitleName: "1AmazingDB",
			},
			expectedErr: errors.New("invalid title: 1AmazingDB. Only alphanumerical characters are allowed (leading character must be a letter)"),
		},
	}

	for _, test := range tests {
		outer.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedErr, test.module.Validate())
		})
	}
}

func TestGenerateWrongModuleName(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	examplesTmp := filepath.Join(tmpCtx.RootDir, "examples")
	examplesDocTmp := filepath.Join(tmpCtx.DocsDir(), "examples")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(examplesTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	require.NoError(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	require.NoError(t, err)

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
		module := context.TestcontainersModule{
			Name:  test.name,
			Image: "docker.io/example/" + test.name + ":latest",
		}

		err = internal.GenerateFiles(tmpCtx, module)
		require.Error(t, err)
	}
}

func TestGenerateWrongModuleTitle(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	examplesTmp := filepath.Join(tmpCtx.RootDir, "examples")
	examplesDocTmp := filepath.Join(tmpCtx.DocsDir(), "examples")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(examplesTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	require.NoError(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	require.NoError(t, err)

	tests := []struct {
		title       string
		expectError bool
	}{
		{title: " fooDB", expectError: true},
		{title: "fooDB ", expectError: true},
		{title: "foo barDB", expectError: true},
		{title: "foo-barDB", expectError: true},
		{title: "foo/barDB", expectError: true},
		{title: "foo\\barDB", expectError: true},
		{title: "1fooDB", expectError: true},
		{title: "foo1DB", expectError: false},
		{title: "-fooDB", expectError: true},
		{title: "foo-DB", expectError: true},
	}

	for _, test := range tests {
		module := context.TestcontainersModule{
			Name:      "foo",
			TitleName: test.title,
			Image:     "docker.io/example/foo:latest",
		}

		err = internal.GenerateFiles(tmpCtx, module)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestGenerate(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	examplesTmp := filepath.Join(tmpCtx.RootDir, "examples")
	examplesDocTmp := filepath.Join(tmpCtx.DocsDir(), "examples")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(examplesTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	require.NoError(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	require.NoError(t, err)

	originalConfig, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	require.NoError(t, err)

	module := context.TestcontainersModule{
		Name:      "foodb4tw",
		TitleName: "FooDB4TheWin",
		IsModule:  false,
		Image:     "docker.io/example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err = internal.GenerateFiles(tmpCtx, module)
	require.NoError(t, err)

	moduleDirPath := filepath.Join(examplesTmp, moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	require.NoError(t, err) // error nil implies the file exist
	assert.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(examplesDocTmp, moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	require.NoError(t, err) // error nil implies the file exist

	mainWorkflowFile := filepath.Join(githubWorkflowsTmp, "ci.yml")
	_, err = os.Stat(mainWorkflowFile)
	require.NoError(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)
	assertModuleGithubWorkflowContent(t, mainWorkflowFile)

	generatedTemplatesDir := filepath.Join(examplesTmp, moduleNameLower)
	// do not generate examples_test.go for examples
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, module, originalConfig, tmpCtx)
}

func TestGenerateModule(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	modulesTmp := filepath.Join(tmpCtx.RootDir, "modules")
	modulesDocTmp := filepath.Join(tmpCtx.DocsDir(), "modules")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(modulesTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(modulesDocTmp, 0o777)
	require.NoError(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	require.NoError(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	require.NoError(t, err)

	originalConfig, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	require.NoError(t, err)

	module := context.TestcontainersModule{
		Name:      "foodb",
		TitleName: "FooDB",
		IsModule:  true,
		Image:     "docker.io/example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err = internal.GenerateFiles(tmpCtx, module)
	require.NoError(t, err)

	moduleDirPath := filepath.Join(modulesTmp, moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	require.NoError(t, err) // error nil implies the file exist
	assert.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(modulesDocTmp, moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	require.NoError(t, err) // error nil implies the file exist

	mainWorkflowFile := filepath.Join(githubWorkflowsTmp, "ci.yml")
	_, err = os.Stat(mainWorkflowFile)
	require.NoError(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)
	assertModuleGithubWorkflowContent(t, mainWorkflowFile)

	generatedTemplatesDir := filepath.Join(modulesTmp, moduleNameLower)
	assertExamplesTestContent(t, module, filepath.Join(generatedTemplatesDir, "examples_test.go"))
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, module, originalConfig, tmpCtx)
}

// assert content module file in the docs
func assertModuleDocContent(t *testing.T, module context.TestcontainersModule, moduleDocFile string) {
	content, err := os.ReadFile(moduleDocFile)
	require.NoError(t, err)

	lower := module.Lower()
	title := module.Title()

	data := sanitiseContent(content)
	assert.Equal(t, "# "+title, data[0])
	assert.Equal(t, `Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>`, data[2])
	assert.Equal(t, "## Introduction", data[4])
	assert.Equal(t, "The Testcontainers module for "+title+".", data[6])
	assert.Equal(t, "## Adding this module to your project dependencies", data[8])
	assert.Equal(t, "Please run the following command to add the "+title+" module to your Go dependencies:", data[10])
	assert.Equal(t, "go get github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+lower, data[13])
	assert.Equal(t, "<!--codeinclude-->", data[18])
	assert.Equal(t, "[Creating a "+title+" container](../../"+module.ParentDir()+"/"+lower+"/examples_test.go) inside_block:run"+title+"Container", data[19])
	assert.Equal(t, "<!--/codeinclude-->", data[20])
	assert.Equal(t, "The "+title+" module exposes one entrypoint function to create the "+title+" container, and this function receives three parameters:", data[31])
	assert.True(t, strings.HasSuffix(data[34], "(*"+title+"Container, error)"))
	assert.Equal(t, "If you need to set a different "+title+" Docker image, you can set a valid Docker image as the second argument in the `Run` function.", data[47])
	assert.Equal(t, "E.g. `Run(context.Background(), \""+module.Image+"\")`.", data[48])
}

// assert content module test
func assertExamplesTestContent(t *testing.T, module context.TestcontainersModule, examplesTestFile string) {
	content, err := os.ReadFile(examplesTestFile)
	require.NoError(t, err)

	lower := module.Lower()
	entrypoint := module.Entrypoint()
	title := module.Title()

	data := sanitiseContent(content)
	assert.Equal(t, "package "+lower+"_test", data[0])
	assert.Equal(t, "\t\"github.com/testcontainers/testcontainers-go/modules/"+lower+"\"", data[7])
	assert.Equal(t, "func Example"+entrypoint+"() {", data[10])
	assert.Equal(t, "\t// run"+title+"Container {", data[11])
	assert.Equal(t, "\t"+lower+"Container, err := "+lower+"."+entrypoint+"(ctx, \""+module.Image+"\")", data[14])
	assert.Equal(t, "\tfmt.Println(state.Running)", data[32])
	assert.Equal(t, "\t// Output:", data[34])
	assert.Equal(t, "\t// true", data[35])
}

// assert content module test
func assertModuleTestContent(t *testing.T, module context.TestcontainersModule, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, "package "+module.Lower()+"_test", data[0])
	assert.Equal(t, "func Test"+module.Title()+"(t *testing.T) {", data[11])
	assert.Equal(t, "\tctr, err := "+module.Lower()+"."+module.Entrypoint()+"(ctx, \""+module.Image+"\")", data[14])
}

// assert content module
func assertModuleContent(t *testing.T, module context.TestcontainersModule, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	require.NoError(t, err)

	lower := module.Lower()
	containerName := module.ContainerName()
	exampleName := module.Title()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	require.Equal(t, "package "+lower, data[0])
	require.Equal(t, "// "+containerName+" represents the "+exampleName+" container type used in the module", data[9])
	require.Equal(t, "type "+containerName+" struct {", data[10])
	require.Equal(t, "// "+entrypoint+" creates an instance of the "+exampleName+" container type", data[14])
	require.Equal(t, "func "+entrypoint+"(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*"+containerName+", error) {", data[15])
	require.Equal(t, "\t\tImage: img,", data[17])
	require.Equal(t, "\t\tif err := opt.Customize(&genericContainerReq); err != nil {", data[26])
	require.Equal(t, "\t\t\treturn nil, fmt.Errorf(\"customize: %w\", err)", data[27])
	require.Equal(t, "\tvar c *"+containerName, data[32])
	require.Equal(t, "\t\tc = &"+containerName+"{Container: container}", data[34])
	require.Equal(t, "\treturn c, nil", data[41])
}

// assert content GitHub workflow for the module
func assertModuleGithubWorkflowContent(t *testing.T, moduleWorkflowFile string) {
	content, err := os.ReadFile(moduleWorkflowFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	ctx := getTestRootContext(t)

	modulesList, err := ctx.GetModules()
	require.NoError(t, err)
	assert.Equal(t, "        module: ["+strings.Join(modulesList, ", ")+"]", data[96])

	examplesList, err := ctx.GetExamples()
	require.NoError(t, err)
	assert.Equal(t, "        module: ["+strings.Join(examplesList, ", ")+"]", data[111])
}

// assert content go.mod
func assertGoModContent(t *testing.T, module context.TestcontainersModule, tcVersion string, goModFile string) {
	content, err := os.ReadFile(goModFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, "module github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+module.Lower(), data[0])
	assert.Equal(t, "require github.com/testcontainers/testcontainers-go "+tcVersion, data[4])
	assert.Equal(t, "replace github.com/testcontainers/testcontainers-go => ../..", data[6])
}

// assert content Makefile
func assertMakefileContent(t *testing.T, module context.TestcontainersModule, makefile string) {
	content, err := os.ReadFile(makefile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, data[4], "\t$(MAKE) test-"+module.Lower())
}

// assert content in the nav items from mkdocs.yml
func assertMkdocsNavItems(t *testing.T, module context.TestcontainersModule, originalConfig *mkdocs.Config, tmpCtx context.Context) {
	config, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	require.NoError(t, err)

	parentDir := module.ParentDir()

	navItems := config.Nav[4].Examples
	expectedEntries := originalConfig.Nav[4].Examples
	if module.IsModule {
		navItems = config.Nav[3].Modules
		expectedEntries = originalConfig.Nav[3].Modules
	}

	assert.Len(t, navItems, len(expectedEntries)+1)

	// the module should be in the nav
	found := false
	for _, ex := range navItems {
		markdownModule := module.ParentDir() + "/" + module.Lower() + ".md"
		if markdownModule == ex {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the index
	assert.Equal(t, parentDir+"/index.md", navItems[0], navItems)
}

func sanitiseContent(bytes []byte) []string {
	content := string(bytes)

	// Windows uses \r\n for newlines, but we want to use \n
	content = strings.ReplaceAll(content, "\r\n", "\n")

	data := strings.Split(content, "\n")

	return data
}
