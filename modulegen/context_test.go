package main_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	main "github.com/testcontainers/testcontainers-go/modulegen"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
)

func TestGetDependabotConfigFile(t *testing.T) {
	tmp := t.TempDir()

	ctx := &main.Context{RootDir: filepath.Join(tmp, "testcontainers-go")}

	githubDir := ctx.GithubDir()
	cfgFile := ctx.DependabotConfigFile()
	err := os.MkdirAll(githubDir, 0o777)
	require.NoError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0o777)
	require.NoError(t, err)

	file := ctx.DependabotConfigFile()
	require.NotNil(t, file)

	assert.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", ".github", "dependabot.yml")))
}

func TestExamplesHasDependabotEntry(t *testing.T) {
	rootDir, err := getRootDir()
	require.NoError(t, err)
	ctx := &main.Context{RootDir: rootDir}
	examples, err := ctx.GetExamples()
	require.NoError(t, err)
	dependabotUpdates, err := dependabot.GetUpdates(ctx.DependabotConfigFile())
	require.NoError(t, err)

	exampleUpdates := []dependabot.Update{}
	// exclude the Go modules from the examples updates
	for _, update := range dependabotUpdates {
		if strings.HasPrefix(update.Directory, "/examples/") {
			exampleUpdates = append(exampleUpdates, update)
		}
	}

	assert.Equal(t, len(exampleUpdates), len(examples))

	// all example modules exist in the dependabot updates
	for _, example := range examples {
		found := false
		for _, exampleUpdate := range exampleUpdates {
			dependabotDir := "/examples/" + example

			assert.Equal(t, exampleUpdate.Schedule.Interval, "monthly")

			if dependabotDir == exampleUpdate.Directory {
				found = true
				continue
			}
		}
		assert.True(t, found, "example %s is not present in the dependabot updates", example)
	}
}

func TestModulesHasDependabotEntry(t *testing.T) {
	rootDir, err := getRootDir()
	require.NoError(t, err)
	ctx := &main.Context{RootDir: rootDir}
	modules, err := ctx.GetModules()
	require.NoError(t, err)
	dependabotUpdates, err := dependabot.GetUpdates(ctx.DependabotConfigFile())
	require.NoError(t, err)

	moduleUpdates := []dependabot.Update{}
	// exclude the Go modules from the examples updates
	for _, update := range dependabotUpdates {
		if strings.HasPrefix(update.Directory, "/modules/") {
			moduleUpdates = append(moduleUpdates, update)
		}
	}
	assert.Equal(t, len(moduleUpdates), len(modules))

	// all module modules exist in the dependabot updates
	for _, module := range modules {
		found := false
		for _, moduleUpdate := range moduleUpdates {
			dependabotDir := "/modules/" + module

			assert.Equal(t, moduleUpdate.Schedule.Interval, "monthly")

			if dependabotDir == moduleUpdate.Directory {
				found = true
				continue
			}
		}
		assert.True(t, found, "module %s is not present in the dependabot updates", module)
	}
}

func getRootDir() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Dir(current), nil
}
