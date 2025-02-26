package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/dependabot"
)

func TestGetDependabotConfigFile(t *testing.T) {
	ctx := context.New(filepath.Join(t.TempDir(), "testcontainers-go"))

	githubDir := ctx.GithubDir()
	cfgFile := ctx.DependabotConfigFile()
	err := os.MkdirAll(githubDir, 0o777)
	require.NoError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0o777)
	require.NoError(t, err)

	file := ctx.DependabotConfigFile()
	require.NotNil(t, file)

	require.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", ".github", "dependabot.yml")))
}

func TestExamplesHasDependabotEntry(t *testing.T) {
	testProject := copyInitialProject(t)

	examples, err := testProject.ctx.GetExamples()
	require.NoError(t, err)
	dependabotUpdates, err := dependabot.GetUpdates(testProject.ctx.DependabotConfigFile())
	require.NoError(t, err)

	exampleUpdates := []dependabot.Update{}
	// exclude the Go modules from the examples updates
	for _, update := range dependabotUpdates {
		if strings.HasPrefix(update.Directory, "/examples/") {
			exampleUpdates = append(exampleUpdates, update)
		}
	}

	require.Equal(t, len(exampleUpdates), len(examples))

	// all example modules exist in the dependabot updates
	for _, example := range examples {
		found := false
		for _, exampleUpdate := range exampleUpdates {
			dependabotDir := "/examples/" + example

			require.Equal(t, "monthly", exampleUpdate.Schedule.Interval)

			if dependabotDir == exampleUpdate.Directory {
				found = true
				continue
			}
		}
		require.True(t, found, "example %s is not present in the dependabot updates", example)
	}
}

func TestModulesHasDependabotEntry(t *testing.T) {
	testProject := copyInitialProject(t)

	modules, err := testProject.ctx.GetModules()
	require.NoError(t, err)
	dependabotUpdates, err := dependabot.GetUpdates(testProject.ctx.DependabotConfigFile())
	require.NoError(t, err)

	moduleUpdates := []dependabot.Update{}
	// exclude the Go modules from the examples updates
	for _, update := range dependabotUpdates {
		if strings.HasPrefix(update.Directory, "/modules/") {
			moduleUpdates = append(moduleUpdates, update)
		}
	}
	require.Equal(t, len(moduleUpdates), len(modules))

	// all module modules exist in the dependabot updates
	for _, module := range modules {
		found := false
		for _, moduleUpdate := range moduleUpdates {
			dependabotDir := "/modules/" + module

			require.Equal(t, "monthly", moduleUpdate.Schedule.Interval)

			if dependabotDir == moduleUpdate.Directory {
				found = true
				continue
			}
		}
		require.True(t, found, "module %s is not present in the dependabot updates", module)
	}
}
