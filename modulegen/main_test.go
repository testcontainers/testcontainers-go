package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
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
			expectedEntrypoint:    "RunContainer",
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
			expectedEntrypoint:    "RunContainer",
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
			expectedEntrypoint:    "runContainer",
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
			expectedEntrypoint:    "runContainer",
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
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
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
		module := context.TestcontainersModule{
			Name:  test.name,
			Image: "docker.io/example/" + test.name + ":latest",
		}

		err = internal.GenerateFiles(tmpCtx, module)
		assert.Error(t, err)
	}
}

func TestGenerateWrongModuleTitle(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	examplesTmp := filepath.Join(tmpCtx.RootDir, "examples")
	examplesDocTmp := filepath.Join(tmpCtx.DocsDir(), "examples")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(examplesTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
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
		module := context.TestcontainersModule{
			Name:      "foo",
			TitleName: test.title,
			Image:     "docker.io/example/foo:latest",
		}

		err = internal.GenerateFiles(tmpCtx, module)
		assert.Error(t, err)
	}
}

func TestGenerate(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	examplesTmp := filepath.Join(tmpCtx.RootDir, "examples")
	examplesDocTmp := filepath.Join(tmpCtx.DocsDir(), "examples")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(examplesTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(examplesDocTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	assert.Nil(t, err)

	originalConfig, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	assert.Nil(t, err)

	err = copyInitialDependabotConfig(t, tmpCtx)
	assert.Nil(t, err)

	originalDependabotConfigUpdates, err := dependabot.GetUpdates(tmpCtx.DependabotConfigFile())
	assert.Nil(t, err)

	module := context.TestcontainersModule{
		Name:      "foodb4tw",
		TitleName: "FooDB4TheWin",
		IsModule:  false,
		Image:     "docker.io/example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err = internal.GenerateFiles(tmpCtx, module)
	assert.Nil(t, err)

	moduleDirPath := filepath.Join(examplesTmp, moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	assert.Nil(t, err) // error nil implies the file exist
	assert.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(examplesDocTmp, moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	assert.Nil(t, err) // error nil implies the file exist

	mainWorkflowFile := filepath.Join(githubWorkflowsTmp, "ci.yml")
	_, err = os.Stat(mainWorkflowFile)
	assert.Nil(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)
	assertModuleGithubWorkflowContent(t, module, mainWorkflowFile)

	generatedTemplatesDir := filepath.Join(examplesTmp, moduleNameLower)
	// do not generate examples_test.go for examples
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, module, originalConfig, tmpCtx)
	assertDependabotUpdates(t, module, originalDependabotConfigUpdates, tmpCtx)
}

func TestGenerateModule(t *testing.T) {
	tmpCtx := context.New(t.TempDir())
	modulesTmp := filepath.Join(tmpCtx.RootDir, "modules")
	modulesDocTmp := filepath.Join(tmpCtx.DocsDir(), "modules")
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()

	err := os.MkdirAll(modulesTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(modulesDocTmp, 0o777)
	assert.Nil(t, err)
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	assert.Nil(t, err)

	err = copyInitialMkdocsConfig(t, tmpCtx)
	assert.Nil(t, err)

	originalConfig, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	assert.Nil(t, err)

	err = copyInitialDependabotConfig(t, tmpCtx)
	assert.Nil(t, err)

	originalDependabotConfigUpdates, err := dependabot.GetUpdates(tmpCtx.DependabotConfigFile())
	assert.Nil(t, err)

	module := context.TestcontainersModule{
		Name:      "foodb",
		TitleName: "FooDB",
		IsModule:  true,
		Image:     "docker.io/example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err = internal.GenerateFiles(tmpCtx, module)
	assert.Nil(t, err)

	moduleDirPath := filepath.Join(modulesTmp, moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	assert.Nil(t, err) // error nil implies the file exist
	assert.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(modulesDocTmp, moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	assert.Nil(t, err) // error nil implies the file exist

	mainWorkflowFile := filepath.Join(githubWorkflowsTmp, "ci.yml")
	_, err = os.Stat(mainWorkflowFile)
	assert.Nil(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)
	assertModuleGithubWorkflowContent(t, module, mainWorkflowFile)

	generatedTemplatesDir := filepath.Join(modulesTmp, moduleNameLower)
	assertExamplesTestContent(t, module, filepath.Join(generatedTemplatesDir, "examples_test.go"))
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, module, originalConfig, tmpCtx)
	assertDependabotUpdates(t, module, originalDependabotConfigUpdates, tmpCtx)
}

// assert content in the Dependabot descriptor file
func assertDependabotUpdates(t *testing.T, module context.TestcontainersModule, originalConfigUpdates dependabot.Updates, tmpCtx context.Context) {
	modules, err := dependabot.GetUpdates(tmpCtx.DependabotConfigFile())
	assert.Nil(t, err)

	assert.Equal(t, len(originalConfigUpdates)+1, len(modules))

	// the module should be in the dependabot updates
	found := false
	for _, ex := range modules {
		directory := "/" + module.ParentDir() + "/" + module.Lower()
		if directory == ex.Directory {
			found = true
		}
	}

	assert.True(t, found)

	// first item is the github-actions module
	assert.Equal(t, "/", modules[0].Directory, modules)
	assert.Equal(t, "github-actions", modules[0].PackageEcosystem, "PackageEcosystem should be github-actions")

	// second item is the core module
	assert.Equal(t, "/", modules[1].Directory, modules)
	assert.Equal(t, "gomod", modules[1].PackageEcosystem, "PackageEcosystem should be gomod")

	// third item is the pip module
	assert.Equal(t, "/", modules[2].Directory, modules)
	assert.Equal(t, "pip", modules[2].PackageEcosystem, "PackageEcosystem should be pip")
}

// assert content module file in the docs
func assertModuleDocContent(t *testing.T, module context.TestcontainersModule, moduleDocFile string) {
	content, err := os.ReadFile(moduleDocFile)
	assert.Nil(t, err)

	lower := module.Lower()
	title := module.Title()

	data := sanitiseContent(content)
	assert.Equal(t, data[0], "# "+title)
	assert.Equal(t, data[2], `Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>`)
	assert.Equal(t, data[4], "## Introduction")
	assert.Equal(t, data[6], "The Testcontainers module for "+title+".")
	assert.Equal(t, data[8], "## Adding this module to your project dependencies")
	assert.Equal(t, data[10], "Please run the following command to add the "+title+" module to your Go dependencies:")
	assert.Equal(t, data[13], "go get github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+lower)
	assert.Equal(t, data[18], "<!--codeinclude-->")
	assert.Equal(t, data[19], "[Creating a "+title+" container](../../"+module.ParentDir()+"/"+lower+"/examples_test.go) inside_block:run"+title+"Container")
	assert.Equal(t, data[20], "<!--/codeinclude-->")
	assert.Equal(t, data[24], "The "+title+" module exposes one entrypoint function to create the "+title+" container, and this function receives two parameters:")
	assert.True(t, strings.HasSuffix(data[27], "(*"+title+"Container, error)"))
	assert.Equal(t, "for "+title+". E.g. `testcontainers.WithImage(\""+module.Image+"\")`.", data[40])
}

// assert content module test
func assertExamplesTestContent(t *testing.T, module context.TestcontainersModule, examplesTestFile string) {
	content, err := os.ReadFile(examplesTestFile)
	assert.Nil(t, err)

	lower := module.Lower()
	entrypoint := module.Entrypoint()
	title := module.Title()

	data := sanitiseContent(content)
	assert.Equal(t, data[0], "package "+lower+"_test")
	assert.Equal(t, data[6], "\t\"github.com/testcontainers/testcontainers-go\"")
	assert.Equal(t, data[7], "\t\"github.com/testcontainers/testcontainers-go/modules/"+lower+"\"")
	assert.Equal(t, data[10], "func Example"+entrypoint+"() {")
	assert.Equal(t, data[11], "\t// run"+title+"Container {")
	assert.Equal(t, data[14], "\t"+lower+"Container, err := "+lower+"."+entrypoint+"(ctx, testcontainers.WithImage(\""+module.Image+"\"))")
	assert.Equal(t, data[32], "\tfmt.Println(state.Running)")
	assert.Equal(t, data[34], "\t// Output:")
	assert.Equal(t, data[35], "\t// true")
}

// assert content module test
func assertModuleTestContent(t *testing.T, module context.TestcontainersModule, exampleTestFile string) {
	content, err := os.ReadFile(exampleTestFile)
	assert.Nil(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, data[0], "package "+module.Lower())
	assert.Equal(t, data[9], "func Test"+module.Title()+"(t *testing.T) {")
	assert.Equal(t, data[12], "\tcontainer, err := "+module.Entrypoint()+"(ctx, testcontainers.WithImage(\""+module.Image+"\"))")
}

// assert content module
func assertModuleContent(t *testing.T, module context.TestcontainersModule, exampleFile string) {
	content, err := os.ReadFile(exampleFile)
	assert.Nil(t, err)

	lower := module.Lower()
	containerName := module.ContainerName()
	exampleName := module.Title()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	assert.Equal(t, data[0], "package "+lower)
	assert.Equal(t, data[8], "// "+containerName+" represents the "+exampleName+" container type used in the module")
	assert.Equal(t, data[9], "type "+containerName+" struct {")
	assert.Equal(t, data[13], "// "+entrypoint+" creates an instance of the "+exampleName+" container type")
	assert.Equal(t, data[14], "func "+entrypoint+"(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*"+containerName+", error) {")
	assert.Equal(t, data[16], "\t\tImage: \""+module.Image+"\",")
	assert.Equal(t, data[33], "\treturn &"+containerName+"{Container: container}, nil")
}

// assert content GitHub workflow for the module
func assertModuleGithubWorkflowContent(t *testing.T, module context.TestcontainersModule, moduleWorkflowFile string) {
	content, err := os.ReadFile(moduleWorkflowFile)
	assert.Nil(t, err)

	data := sanitiseContent(content)
	ctx := getTestRootContext(t)

	modulesList, err := ctx.GetModules()
	assert.Nil(t, err)
	assert.Equal(t, "        module: ["+strings.Join(modulesList, ", ")+"]", data[108])

	examplesList, err := ctx.GetExamples()
	assert.Nil(t, err)
	assert.Equal(t, "        module: ["+strings.Join(examplesList, ", ")+"]", data[128])
}

// assert content go.mod
func assertGoModContent(t *testing.T, module context.TestcontainersModule, tcVersion string, goModFile string) {
	content, err := os.ReadFile(goModFile)
	assert.Nil(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, "module github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+module.Lower(), data[0])
	assert.Equal(t, "require github.com/testcontainers/testcontainers-go "+tcVersion, data[4])
	assert.Equal(t, "replace github.com/testcontainers/testcontainers-go => ../..", data[6])
}

// assert content Makefile
func assertMakefileContent(t *testing.T, module context.TestcontainersModule, makefile string) {
	content, err := os.ReadFile(makefile)
	assert.Nil(t, err)

	data := sanitiseContent(content)
	assert.Equal(t, data[4], "\t$(MAKE) test-"+module.Lower())
}

// assert content in the nav items from mkdocs.yml
func assertMkdocsNavItems(t *testing.T, module context.TestcontainersModule, originalConfig *mkdocs.Config, tmpCtx context.Context) {
	config, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	assert.Nil(t, err)

	parentDir := module.ParentDir()

	navItems := config.Nav[4].Examples
	expectedEntries := originalConfig.Nav[4].Examples
	if module.IsModule {
		navItems = config.Nav[3].Modules
		expectedEntries = originalConfig.Nav[3].Modules
	}

	assert.Equal(t, len(expectedEntries)+1, len(navItems))

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

func copyInitialDependabotConfig(t *testing.T, tmpCtx context.Context) error {
	ctx := getTestRootContext(t)
	return dependabot.CopyConfig(ctx.DependabotConfigFile(), tmpCtx.DependabotConfigFile())
}
