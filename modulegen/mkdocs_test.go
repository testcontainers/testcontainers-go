package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMkDocsConfigFile(t *testing.T) {
	tmp := t.TempDir()

	rootDir := filepath.Join(tmp, "testcontainers-go")
	cfgFile := filepath.Join(rootDir, "mkdocs.yml")
	err := os.MkdirAll(rootDir, 0777)
	require.NoError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0777)
	require.NoError(t, err)

	file := getMkdocsConfigFile(rootDir)
	require.NotNil(t, file)

	assert.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", "mkdocs.yml")))
}

func TestReadMkDocsConfig(t *testing.T) {
	tmp := t.TempDir()

	rootDir := filepath.Join(tmp, "testcontainers-go")
	err := os.MkdirAll(rootDir, 0777)
	require.NoError(t, err)

	err = copyInitialMkdocsConfig(t, rootDir)
	require.NoError(t, err)

	config, err := readMkdocsConfig(rootDir)
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "Testcontainers for Go", config.SiteName)
	assert.Equal(t, "https://github.com/testcontainers/testcontainers-go", config.RepoURL)
	assert.Equal(t, "edit/main/docs/", config.EditURI)

	// theme
	theme := config.Theme
	assert.Equal(t, "material", theme.Name)

	// nav bar
	nav := config.Nav
	assert.Equal(t, "index.md", nav[0].Home)
	assert.Greater(t, len(nav[2].Features), 0)
	assert.Greater(t, len(nav[3].Modules), 0)
	assert.Greater(t, len(nav[4].Examples), 0)
}

func TestExamples(t *testing.T) {
	examples, err := getExamples()
	require.NoError(t, err)
	examplesDocs, err := getExamplesDocs()
	require.NoError(t, err)

	// we have to remove the index.md file from the examples docs
	assert.Equal(t, len(examplesDocs)-1, len(examples))

	// all example modules exist in the documentation
	for _, example := range examples {
		found := false
		for _, exampleDoc := range examplesDocs {
			markdownName := example.Name() + ".md"

			if markdownName == exampleDoc.Name() {
				found = true
				continue
			}
		}
		assert.True(t, found, "example %s is not present in the docs", example.Name())
	}
}

func copyInitialMkdocsConfig(t *testing.T, tmpDir string) error {
	projectDir, err := getRootDir()
	require.NoError(t, err)

	initialConfig, err := readMkdocsConfig(projectDir)
	require.NoError(t, err)

	return writeMkdocsConfig(tmpDir, initialConfig)
}
