package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
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

			require.Equal(t, "mongodb", module.Lower())
			require.Equal(t, test.expectedTitle, module.Title())
			require.Equal(t, "Container", module.ContainerName())
			require.Equal(t, test.expectedEntrypoint, module.Entrypoint())
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
			if test.expectedErr != nil {
				require.EqualError(t, test.module.Validate(), test.expectedErr.Error())
			} else {
				require.NoError(t, test.module.Validate())
			}
		})
	}
}

func TestGenerateWrongModuleName(t *testing.T) {
	testProject := copyInitialProject(t)

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
			Image: "example/" + test.name + ":latest",
		}

		err := internal.GenerateFiles(testProject.ctx, module)
		require.Error(t, err)
	}
}

func TestGenerateWrongModuleTitle(t *testing.T) {
	testProject := copyInitialProject(t)

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
			Image:     "example/foo:latest",
		}

		err := internal.GenerateFiles(testProject.ctx, module)
		if test.expectError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestGenerate(t *testing.T) {
	testProject := copyInitialProject(t)

	module := context.TestcontainersModule{
		Name:      "foodb4tw",
		TitleName: "FooDB4TheWin",
		IsModule:  false,
		Image:     "example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err := internal.GenerateFiles(testProject.ctx, module)
	require.NoError(t, err)

	moduleDirPath := filepath.Join(testProject.ctx.ExamplesDir(), moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	require.NoError(t, err) // error nil implies the file exist
	require.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(testProject.ctx.ExamplesDocsDir(), moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	require.NoError(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)

	generatedTemplatesDir := filepath.Join(testProject.ctx.ExamplesDir(), moduleNameLower)
	// do not generate examples_test.go for examples
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, testProject.originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, testProject.ctx, module)
	assertDependabotUpdates(t, testProject.ctx, module)
}

func TestGenerateModule(t *testing.T) {
	testProject := copyInitialProject(t)

	module := context.TestcontainersModule{
		Name:      "foodb",
		TitleName: "FooDB",
		IsModule:  true,
		Image:     "example/foodb:latest",
	}
	moduleNameLower := module.Lower()

	err := internal.GenerateFiles(testProject.ctx, module)
	require.NoError(t, err)

	moduleDirPath := filepath.Join(testProject.ctx.ModulesDir(), moduleNameLower)

	moduleDirFileInfo, err := os.Stat(moduleDirPath)
	require.NoError(t, err) // error nil implies the file exist
	require.True(t, moduleDirFileInfo.IsDir())

	moduleDocFile := filepath.Join(testProject.ctx.ModulesDocsDir(), moduleNameLower+".md")
	_, err = os.Stat(moduleDocFile)
	require.NoError(t, err) // error nil implies the file exist

	assertModuleDocContent(t, module, moduleDocFile)

	generatedTemplatesDir := filepath.Join(testProject.ctx.ModulesDir(), moduleNameLower)
	assertExamplesTestContent(t, module, filepath.Join(generatedTemplatesDir, "examples_test.go"))
	assertModuleTestContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+"_test.go"))
	assertModuleContent(t, module, filepath.Join(generatedTemplatesDir, moduleNameLower+".go"))
	assertGoModContent(t, module, testProject.originalConfig.Extra.LatestVersion, filepath.Join(generatedTemplatesDir, "go.mod"))
	assertMakefileContent(t, module, filepath.Join(generatedTemplatesDir, "Makefile"))
	assertMkdocsNavItems(t, testProject.ctx, module)
	assertDependabotUpdates(t, testProject.ctx, module)
}

func TestRefresh(t *testing.T) {
	testProject := copyInitialProject(t)

	err := internal.Refresh(testProject.ctx)
	require.NoError(t, err)

	var modulesAndExamples []context.TestcontainersModule

	examples, err := testProject.ctx.GetExamples()
	require.NoError(t, err)

	for _, example := range examples {
		modulesAndExamples = append(modulesAndExamples, context.TestcontainersModule{
			Name:     example,
			IsModule: false,
		})
	}

	modules, err := testProject.ctx.GetModules()
	require.NoError(t, err)

	for _, module := range modules {
		modulesAndExamples = append(modulesAndExamples, context.TestcontainersModule{
			Name:     module,
			IsModule: true,
		})
	}

	for _, module := range modulesAndExamples {
		assertMkdocsNavItems(t, testProject.ctx, module)
		assertDependabotUpdates(t, testProject.ctx, module)
	}
}

// assert content in the Dependabot descriptor file
func assertDependabotUpdates(t *testing.T, tmpCtx context.Context, module context.TestcontainersModule) {
	t.Helper()
	updates, err := dependabot.GetUpdates(tmpCtx.DependabotConfigFile())
	require.NoError(t, err)

	// first item is the github-actions module, which uses an array of directories
	require.Len(t, updates[0].Directories, 1)
	require.Equal(t, "/", updates[0].Directories[0], updates)
	require.Equal(t, "github-actions", updates[0].PackageEcosystem, "PackageEcosystem should be github-actions")

	// second item is the Go modules
	require.Equal(t, "gomod", updates[1].PackageEcosystem, "PackageEcosystem should be gomod")

	// modulegen exists as a gomod module
	require.True(t, slices.Contains(updates[1].Directories, "/modulegen"), "modulegen should exist")

	// the module should be in the dependabot updates in the gomod section
	directory := "/" + module.ParentDir() + "/" + module.Lower()
	require.True(t, slices.Contains(updates[1].Directories, directory), "module should exist")

	// third item is the pip module
	require.Equal(t, "pip", updates[2].PackageEcosystem, "PackageEcosystem should be pip")
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
	require.Equal(t, "# "+title, data[0])
	require.Equal(t, `Not available until the next release of testcontainers-go <a href="https://github.com/testcontainers/testcontainers-go"><span class="tc-version">:material-tag: main</span></a>`, data[2])
	require.Equal(t, "## Introduction", data[4])
	require.Equal(t, "The Testcontainers module for "+title+".", data[6])
	require.Equal(t, "## Adding this module to your project dependencies", data[8])
	require.Equal(t, "Please run the following command to add the "+title+" module to your Go dependencies:", data[10])
	require.Equal(t, "go get github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+lower, data[13])
	require.Equal(t, "<!--codeinclude-->", data[18])
	require.Equal(t, "[Creating a "+title+" container](../../"+module.ParentDir()+"/"+lower+"/examples_test.go) inside_block:Example"+entrypoint, data[19])
	require.Equal(t, "<!--/codeinclude-->", data[20])
	require.Equal(t, "The "+title+" module exposes one entrypoint function to create the "+title+" container, and this function receives three parameters:", data[28])
	require.True(t, strings.HasSuffix(data[31], "(*"+title+"Container, error)"))
	require.Equal(t, "Use the second argument in the `Run` function to set a valid Docker image.", data[44])
	require.Equal(t, "In example: `Run(context.Background(), \""+module.Image+"\")`.", data[45])
}

// assert content module test
func assertExamplesTestContent(t *testing.T, module context.TestcontainersModule, examplesTestFile string) {
	t.Helper()
	content, err := os.ReadFile(examplesTestFile)
	require.NoError(t, err)

	lower := module.Lower()
	entrypoint := module.Entrypoint()

	data := sanitiseContent(content)
	require.Equal(t, "package "+lower+"_test", data[0])
	require.Equal(t, "\t\"github.com/testcontainers/testcontainers-go\"", data[7])
	require.Equal(t, "\t\"github.com/testcontainers/testcontainers-go/modules/"+lower+"\"", data[8])
	require.Equal(t, "func Example"+entrypoint+"() {", data[11])
	require.Equal(t, "\t"+lower+"Container, err := "+lower+"."+entrypoint+"(ctx, \""+module.Image+"\")", data[14])
	require.Equal(t, "\tfmt.Println(state.Running)", data[32])
	require.Equal(t, "\t// Output:", data[34])
	require.Equal(t, "\t// true", data[35])
}

// assert content module test
func assertModuleTestContent(t *testing.T, module context.TestcontainersModule, exampleTestFile string) {
	t.Helper()
	content, err := os.ReadFile(exampleTestFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	require.Equal(t, "package "+module.Lower()+"_test", data[0])
	require.Equal(t, "func Test"+module.Title()+"(t *testing.T) {", data[12])
	require.Equal(t, "\tctr, err := "+module.Lower()+"."+module.Entrypoint()+"(ctx, \""+module.Image+"\")", data[15])
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

// assert content go.mod
func assertGoModContent(t *testing.T, module context.TestcontainersModule, tcVersion string, goModFile string) {
	t.Helper()
	content, err := os.ReadFile(goModFile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	require.Equal(t, "module github.com/testcontainers/testcontainers-go/"+module.ParentDir()+"/"+module.Lower(), data[0])
	require.Equal(t, "require github.com/testcontainers/testcontainers-go "+tcVersion, data[4])
	require.Equal(t, "replace github.com/testcontainers/testcontainers-go => ../..", data[6])
}

// assert content Makefile
func assertMakefileContent(t *testing.T, module context.TestcontainersModule, makefile string) {
	t.Helper()
	content, err := os.ReadFile(makefile)
	require.NoError(t, err)

	data := sanitiseContent(content)
	require.Equal(t, data[4], "\t$(MAKE) test-"+module.Lower())
}

// assert content in the nav items from mkdocs.yml
func assertMkdocsNavItems(t *testing.T, ctx context.Context, module context.TestcontainersModule) {
	t.Helper()

	config, err := mkdocs.ReadConfig(ctx.MkdocsConfigFile())
	require.NoError(t, err)

	parentDir := module.ParentDir()

	navItems := config.Nav[4].Examples
	if module.IsModule {
		navItems = config.Nav[3].Modules
	}

	// the module should be in the nav
	found := false
	for _, ex := range navItems {
		markdownModule := filepath.Join(module.ParentDir(), module.Lower()) + ".md"
		if markdownModule == ex {
			found = true
		}
	}

	// confirm compose is not in the nav
	require.NotContains(t, navItems, "modules/compose")
	if module.Lower() != "compose" {
		require.True(t, found, "module %s not found in nav items", module.Lower())
	}

	// first item is the index
	require.Equal(t, filepath.Join(parentDir, "index.md"), navItems[0], navItems)
}

func sanitiseContent(bytes []byte) []string {
	content := string(bytes)

	// Windows uses \r\n for newlines, but we want to use \n
	content = strings.ReplaceAll(content, "\r\n", "\n")

	data := strings.Split(content, "\n")

	return data
}

type testProject struct {
	ctx            context.Context
	originalConfig *mkdocs.Config
}

func copyInitialProject(t *testing.T) testProject {
	t.Helper()

	tmpCtx := context.New(t.TempDir())

	current, err := os.Getwd()
	require.NoError(t, err)

	// Obtains the root context from the current working directory, its parent,
	// as the modulegen tool lives in the root of the project.
	ctx := context.New(filepath.Dir(current))

	// examples and modules
	moduleTypes := []string{"examples", "modules"}
	for _, moduleType := range moduleTypes {
		moduleTypeDir := filepath.Join(tmpCtx.RootDir, moduleType)
		err := os.MkdirAll(moduleTypeDir, 0o777)
		require.NoError(t, err)

		// just create the moduleTypeDir in the tmp context
		files, err := os.ReadDir(filepath.Join(ctx.RootDir, moduleType))
		require.NoError(t, err)
		for _, f := range files {
			if !f.IsDir() {
				continue
			}

			err = os.MkdirAll(filepath.Join(tmpCtx.RootDir, moduleType, f.Name()), 0o777)
			require.NoError(t, err)
		}

		moduleTypeDocDir := filepath.Join(tmpCtx.DocsDir(), moduleType)
		err = os.MkdirAll(moduleTypeDocDir, 0o777)
		require.NoError(t, err)

		// copy all markdown files from the root context's docs into the tmp context's docs
		mkFiles, err := os.ReadDir(filepath.Join(ctx.DocsDir(), moduleType))
		require.NoError(t, err)
		for _, mkFile := range mkFiles {
			srcReader, err := os.Open(filepath.Join(ctx.DocsDir(), moduleType, mkFile.Name()))
			require.NoError(t, err)
			defer srcReader.Close()

			dstWriter, err := os.Create(filepath.Join(tmpCtx.DocsDir(), moduleType, mkFile.Name()))
			require.NoError(t, err)
			defer dstWriter.Close()

			_, err = io.Copy(dstWriter, srcReader)
			require.NoError(t, err)
		}
	}

	// .github/workflows
	githubWorkflowsTmp := tmpCtx.GithubWorkflowsDir()
	err = os.MkdirAll(githubWorkflowsTmp, 0o777)
	require.NoError(t, err)

	// mkdocs.yml
	err = mkdocs.CopyConfig(ctx.MkdocsConfigFile(), tmpCtx.MkdocsConfigFile())
	require.NoError(t, err)

	originalConfig, err := mkdocs.ReadConfig(tmpCtx.MkdocsConfigFile())
	require.NoError(t, err)

	// .github/dependabot.yml
	err = dependabot.CopyConfig(ctx.DependabotConfigFile(), tmpCtx.DependabotConfigFile())
	require.NoError(t, err)

	// go.mod
	goModFile, err := os.ReadFile(ctx.GoModFile())
	require.NoError(t, err)

	err = os.WriteFile(tmpCtx.GoModFile(), goModFile, 0o777)
	require.NoError(t, err)

	// .vscode/testcontainers-go.code-workspace
	err = os.MkdirAll(filepath.Dir(tmpCtx.VSCodeWorkspaceFile()), 0o777)
	require.NoError(t, err)

	vsCodeWorkspaceFile, err := os.ReadFile(ctx.VSCodeWorkspaceFile())
	require.NoError(t, err)

	err = os.WriteFile(tmpCtx.VSCodeWorkspaceFile(), vsCodeWorkspaceFile, 0o777)
	require.NoError(t, err)

	return testProject{
		ctx:            tmpCtx,
		originalConfig: originalConfig,
	}
}
