package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDependabotConfigFile(t *testing.T) {
	tmp := t.TempDir()

	rootDir := filepath.Join(tmp, "testcontainers-go")
	githubDir := filepath.Join(rootDir, ".github")
	cfgFile := filepath.Join(githubDir, "dependabot.yml")
	err := os.MkdirAll(githubDir, 0777)
	require.NoError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0777)
	require.NoError(t, err)

	file := getDependabotConfigFile(rootDir)
	require.NotNil(t, file)

	assert.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", ".github", "dependabot.yml")))
}

func TestReadDependabotConfig(t *testing.T) {
	tmp := t.TempDir()

	rootDir := filepath.Join(tmp, "testcontainers-go")
	githubDir := filepath.Join(rootDir, ".github")
	err := os.MkdirAll(githubDir, 0777)
	require.NoError(t, err)

	err = copyInitialDependabotConfig(t, rootDir)
	require.NoError(t, err)

	config, err := readDependabotConfig(rootDir)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Greater(t, len(config.Updates), 0)
}

func TestExamplesHasDependabotEntry(t *testing.T) {
	examples, err := getExamples()
	require.NoError(t, err)
	exampleUpdates, err := getDependabotUpdates()
	require.NoError(t, err)

	// we have to exclude the main, compose and e2e modules from the examples updates
	assert.Equal(t, len(exampleUpdates)-3, len(examples))

	// all example modules exist in the dependabot updates
	for _, example := range examples {
		found := false
		for _, exampleUpdate := range exampleUpdates {
			dependabotDir := "/examples/" + strings.ToLower(example.Name())

			if dependabotDir == exampleUpdate.Directory {
				found = true
				continue
			}
		}
		assert.True(t, found, "example %s is not present in the dependabot updates", example.Name())
	}
}

func copyInitialDependabotConfig(t *testing.T, tmpDir string) error {
	projectDir, err := getRootDir()
	require.NoError(t, err)

	initialConfig, err := readDependabotConfig(projectDir)
	require.NoError(t, err)

	return writeDependabotConfig(tmpDir, initialConfig)
}
