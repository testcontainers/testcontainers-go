package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/modulegen/internal/context"
	"github.com/testcontainers/testcontainers-go/modulegen/internal/mkdocs"
)

func TestGetMkDocsConfigFile(t *testing.T) {
	tmpCtx := context.New(filepath.Join(t.TempDir(), "testcontainers-go"))
	cfgFile := tmpCtx.MkdocsConfigFile()
	err := os.MkdirAll(tmpCtx.RootDir, 0o777)
	require.NoError(t, err)

	err = os.WriteFile(cfgFile, []byte{}, 0o777)
	require.NoError(t, err)

	file := tmpCtx.MkdocsConfigFile()
	require.NotNil(t, file)

	require.True(t, strings.HasSuffix(file, filepath.Join("testcontainers-go", "mkdocs.yml")))
}

func TestReadMkDocsConfig(t *testing.T) {
	testProject := copyInitialProject(t)

	config, err := mkdocs.ReadConfig(testProject.ctx.MkdocsConfigFile())
	require.NoError(t, err)
	require.NotNil(t, config)

	require.Equal(t, "Testcontainers for Go", config.SiteName)
	require.Equal(t, "https://github.com/testcontainers/testcontainers-go", config.RepoURL)
	require.Equal(t, "edit/main/docs/", config.EditURI)

	// theme
	theme := config.Theme
	require.Equal(t, "material", theme.Name)

	// nav bar
	nav := config.Nav
	require.Equal(t, "index.md", nav[0].Home)
	require.NotEmpty(t, nav[2].Features)
	require.NotEmpty(t, nav[3].Modules)
}
