package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMkDocsConfigFile(t *testing.T) {
	file, err := getMkdocsConfigFile()
	require.NoError(t, err)
	require.NotNil(t, file)

	assert.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", "mkdocs.yml")))
}

func TestReadMkDocsConfig(t *testing.T) {
	config, err := readMkdocsConfig()
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
}

func TestExamples(t *testing.T) {
	examples, err := getExamples()
	require.NoError(t, err)
	examplesDocs, err := getExamplesDocs()
	require.NoError(t, err)

	assert.Equal(t, len(examplesDocs), len(examples))

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
