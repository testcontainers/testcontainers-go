package main

import (
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
		name               string
		module             context.TestcontainersModule
		expectedEntrypoint string
		expectedTitle      string
	}{
		{
			name: "Module with title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  true,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedEntrypoint: "Run",
			expectedTitle:      "MongoDB",
		},
		{
			name: "Module without title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: true,
				Image:    "mongodb:latest",
			},
			expectedEntrypoint: "Run",
			expectedTitle:      "Mongodb",
		},
		{
			name: "Example with title",
			module: context.TestcontainersModule{
				Name:      "mongoDB",
				IsModule:  false,
				Image:     "mongodb:latest",
				TitleName: "MongoDB",
			},
			expectedEntrypoint: "run",
			expectedTitle:      "MongoDB",
		},
		{
			name: "Example without title",
			module: context.TestcontainersModule{
				Name:     "mongoDB",
				IsModule: false,
				Image:    "mongodb:latest",
			},

			expectedEntrypoint: "run",
			expectedTitle:      "Mongodb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			module := test.module

			assert.Equal(t, "mongodb", module.Lower())
			assert.Equal(t, test.expectedTitle, module.Title())
			assert.Equal(t, "Container", module.ContainerName())
			assert.Equal(t, test.expectedEntrypoint, module.Entrypoint())
		})
	}
}

func TestGenerate(t *testing.T) {
	testGenerateModule := func(t *testing.T, module context.TestcontainersModule) {
		t.Helper()

		tmpCtx := context.New(t.TempDir())

		modulesTmp := filepath.Join(tmpCtx.RootDir, module.ParentDir())
		modulesDocTmp := filepath.Join(tmpCtx.DocsDir(), module.ParentDir())
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

		// copy go.work from the root context to the test context
		copyWorkFile(t, tmpCtx)
		// copy go.mod from the root context to the test context
		copyModFile(t, tmpCtx)

		// copy modules and examples from the root context to the test context
		copyModulesAndExamples(t, tmpCtx)

		moduleNameLower := module.Lower()

		err = internal.GenerateFiles(tmpCtx, module)
		require.NoError(t, err)

		moduleDirPath := filepath.Join(modulesTmp, moduleNameLower)

		moduleDirFileInfo, err := os.Stat(moduleDirPath)
		require.NoError(t, err) // error nil implies the file exist
		require.True(t, moduleDirFileInfo.IsDir())

		moduleDocFile := filepath.Join(modulesDocTmp, moduleNameLower+".md")
		_, err = os.Stat(moduleDocFile)
		require.NoError(t, err) // error nil implies the file exist

		mainWorkflowFile := filepath.Join(githubWorkflowsTmp, "ci.yml")
		_, err = os.Stat(mainWorkflowFile)
		require.NoError(t, err) // error nil implies the file exist

		assertModuleDocContent(t, module, moduleDocFile)
		assertModuleGithubWorkflowContent(t, tmpCtx, mainWorkflowFile)
		assertGoWorkContent(t, module, filepath.Join(tmpCtx.RootDir, "go.work"))

		generatedTemplatesDir := filepath.Join(modulesTmp, moduleNameLower)
		assertExamplesTestContent(t, module, filepath.Join(generatedTemplatesDir, "examples_test.go"))
		assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
		assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
		assertGoModContent(t, module, originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
		assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
		assertMkdocsNavItems(t, module, originalConfig, tmpCtx)
	}

	t.Run("Example", func(t *testing.T) {
		module := context.TestcontainersModule{
			Name:      "foodb4tw",
			TitleName: "FooDB4TheWin",
			IsModule:  false,
			Image:     "example/foodb:latest",
		}

		testGenerateModule(t, module)
	})

	t.Run("Module", func(t *testing.T) {
		module := context.TestcontainersModule{
			Name:      "foodb",
			TitleName: "FooDB",
			IsModule:  true,
			Image:     "module/foodb:latest",
		}

		testGenerateModule(t, module)
	})
}

// assert content module file in the docs
func assertModuleDocContent(t *testing.T, module context.TestcontainersModule, moduleDocFile string) {
	t.Helper()
	content, err := os.ReadFile(moduleDocFile)
	require.NoError(t, err)

	lower := module.Lower()
	title := module.Title()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	assert.Equal(t, "# "+title, data[0])
	assert.Equal(t, `Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>`, data[2])
	assert.Equal(t, "## Introduction", data[4])
	assert.Equal(t, "The Testcontainers module for "+title+".", data[6])
	assert.Equal(t, "## Adding this module to your project dependencies", data[8])
	assert.Equal(t, "Please run the following command to add the "+title+" module to your Go dependencies:", data[10])
	assert.Equal(t, "go get github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+lower, data[13])
	assert.Equal(t, "<!--codeinclude-->", data[18])
	assert.Equal(t, "[Creating a "+title+" container](../../"+module.ParentDir()+"/"+lower+"/examples_test.go) inside_block:Example"+entrypoint, data[19])
	assert.Equal(t, "<!--/codeinclude-->", data[20])
	assert.Equal(t, "The "+title+" module exposes one entrypoint function to create the "+title+" container, and this function receives three parameters:", data[31])
	assert.True(t, strings.HasSuffix(data[34], "(*"+title+"Container, error)"))
	assert.Equal(t, "If you need to set a different "+title+" Docker image, you can set a valid Docker image as the second argument in the `Run` function.", data[47])
	assert.Equal(t, "E.g. `Run(context.Background(), \""+module.Image+"\")`.", data[48])
}

// assert content module test
func assertExamplesTestContent(t *testing.T, module context.TestcontainersModule, examplesTestFile string) {
	t.Helper()

	// examples does not have an examples_test.go file
	if !module.IsModule {
		return
	}

	content, err := os.ReadFile(examplesTestFile)
	require.NoError(t, err)

	lower := module.Lower()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	assert.Equal(t, "package "+lower+"_test", data[0])
	assert.Equal(t, "\t\"github.com/testcontainers/testcontainers-go\"", data[7])
	assert.Equal(t, "\t\"github.com/testcontainers/testcontainers-go/modules/"+lower+"\"", data[8])
	assert.Equal(t, "func Example"+entrypoint+"() {", data[11])
	assert.Equal(t, "\t"+lower+"Container, err := "+lower+"."+entrypoint+"(ctx, \""+module.Image+"\")", data[14])
	assert.Equal(t, "\tfmt.Println(state.Running)", data[32])
	assert.Equal(t, "\t// Output:", data[34])
	assert.Equal(t, "\t// true", data[35])
}

// assert content module test
func assertModuleTestContent(t *testing.T, module context.TestcontainersModule, exampleTestFile string) {
	t.Helper()
	content, err := os.ReadFile(exampleTestFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, "package "+module.Lower()+"_test", data[0])
	assert.Equal(t, "func Test"+module.Title()+"(t *testing.T) {", data[12])
	assert.Equal(t, "\tctr, err := "+module.Lower()+"."+module.Entrypoint()+"(ctx, \""+module.Image+"\")", data[15])
}

// assert content module
func assertModuleContent(t *testing.T, module context.TestcontainersModule, exampleFile string) {
	t.Helper()
	content, err := os.ReadFile(exampleFile)
	require.NoError(t, err)

	lower := module.Lower()
	containerName := module.ContainerName()
	exampleName := module.Title()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	require.Equal(t, "package "+lower, data[0])
	require.Equal(t, "// Container represents the "+exampleName+" container type used in the module", data[9])
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
func assertModuleGithubWorkflowContent(t *testing.T, ctx context.Context, moduleWorkflowFile string) {
	t.Helper()
	content, err := os.ReadFile(moduleWorkflowFile)
	require.NoError(t, err)

	data := sanitiseContent(content)

	modulesList, err := ctx.GetModules()
	require.NoError(t, err)
	assert.Equal(t, "        module: ["+strings.Join(modulesList, ", ")+"]", data[96])

	examplesList, err := ctx.GetExamples()
	require.NoError(t, err)
	assert.Equal(t, "        module: ["+strings.Join(examplesList, ", ")+"]", data[111])
}

// assert content go.mod
func assertGoModContent(t *testing.T, module context.TestcontainersModule, tcVersion string, goModFile string) {
	t.Helper()
	content, err := os.ReadFile(goModFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, "module github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+module.Lower(), data[0])
	assert.Equal(t, "require github.com/testcontainers/testcontainers-go "+tcVersion, data[4])
	assert.Equal(t, "replace github.com/testcontainers/testcontainers-go => ../..", data[6])
}

// assert content go.mod
func assertGoWorkContent(t *testing.T, module context.TestcontainersModule, goWorkFile string) {
	t.Helper()
	content, err := os.ReadFile(goWorkFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	require.Contains(t, data, "\t./"+module.ParentDir()+"/"+module.Lower())
}

// assert content Makefile
func assertMakefileContent(t *testing.T, module context.TestcontainersModule, makefile string) {
	t.Helper()
	content, err := os.ReadFile(makefile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, data[4], "\t$(MAKE) test-"+module.Lower())
}

// assert content in the nav items from mkdocs.yml
func assertMkdocsNavItems(t *testing.T, module context.TestcontainersModule, originalConfig *mkdocs.Config, tmpCtx context.Context) {
	t.Helper()
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
