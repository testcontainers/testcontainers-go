package main

import (
	"os"
	"path/filepath"
	"slices"
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

	// should be the second item in the updates
	gomodUpdate := dependabotUpdates[1]

	require.Equal(t, "monthly", gomodUpdate.Schedule.Interval)

	// all example modules exist in the dependabot updates
	for _, example := range examples {
		dependabotDir := "/examples/" + example
		require.True(t, slices.Contains(gomodUpdate.Directories, dependabotDir), "example %s is not present in the dependabot updates", example)
	}
}

func TestModulesHasDependabotEntry(t *testing.T) {
	testProject := copyInitialProject(t)

	modules, err := testProject.ctx.GetModules()
	require.NoError(t, err)
	dependabotUpdates, err := dependabot.GetUpdates(testProject.ctx.DependabotConfigFile())
	require.NoError(t, err)

	// should be the second item in the updates
	gomodUpdate := dependabotUpdates[1]

	require.Equal(t, "monthly", gomodUpdate.Schedule.Interval)

	// all modules exist in the dependabot updates
	for _, module := range modules {
		dependabotDir := "/modules/" + module
		require.True(t, slices.Contains(gomodUpdate.Directories, dependabotDir), "module %s is not present in the dependabot updates", module)
	}
}
