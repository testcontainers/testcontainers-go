package context

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go/internal/docker/config"
	"github.com/testcontainers/testcontainers-go/internal/docker/context/internal"
)

func TestCurrent(t *testing.T) {
	t.Run("current/1", func(tt *testing.T) {
		setupDockerContexts(tt, 1, 3) // current context is context1

		current, err := Current()
		require.NoError(t, err)
		require.Equal(t, "context1", current)
	})

	t.Run("current/auth-error", func(tt *testing.T) {
		tt.Setenv("DOCKER_AUTH_CONFIG", "invalid-auth-config")

		current, err := Current()
		require.Error(t, err)
		require.Empty(t, current)
	})

	t.Run("current/override-host", func(tt *testing.T) {
		tt.Setenv("DOCKER_HOST", "tcp://127.0.0.1:2")

		current, err := Current()
		require.NoError(t, err)
		require.Equal(t, DefaultContextName, current)
	})

	t.Run("current/override-context", func(tt *testing.T) {
		setupDockerContexts(tt, 1, 3)           // current context is context1
		tt.Setenv("DOCKER_CONTEXT", "context2") // override the current context

		current, err := Current()
		require.NoError(t, err)
		require.Equal(t, "context2", current)
	})

	t.Run("current/empty-context", func(tt *testing.T) {
		contextCount := 3
		setupDockerContexts(tt, contextCount+1, contextCount) // current context is the empty one

		current, err := Current()
		require.NoError(t, err)
		require.Equal(t, DefaultContextName, current)
	})
}

func TestCurrentDockerHost(t *testing.T) {
	t.Run("docker-context/1", func(tt *testing.T) {
		setupDockerContexts(tt, 1, 3) // current context is context1

		host, err := CurrentDockerHost()
		require.NoError(t, err)
		require.Equal(t, "tcp://127.0.0.1:1", host) // from context1
	})

	t.Run("docker-context/2", func(tt *testing.T) {
		setupDockerContexts(tt, 2, 3) // current context is context2

		host, err := CurrentDockerHost()
		require.NoError(t, err)
		require.Equal(t, "tcp://127.0.0.1:2", host) // from context2
	})

	t.Run("docker-context/not-found", func(tt *testing.T) {
		setupDockerContexts(tt, 1, 1) // current context is context1

		metaRoot, err := metaRoot()
		require.NoError(t, err)

		host, err := internal.ExtractDockerHost("context-not-found", metaRoot)
		require.Error(t, err)
		require.Empty(t, host)
	})
}

// setupDockerContexts creates a temporary directory structure for testing the Docker context functions.
// It creates the following structure, where $i is the index of the context, starting from 1:
// - $HOME/.docker
//   - config.json
//   - contexts
//   - meta
//   - context$i
//   - meta.json
//
// The config.json file contains the current context, and the meta.json files contain the metadata for each context.
// It generates the specified number of contexts, setting the current context to the one specified by currentContextIndex.
// Finally it always adds a context with an empty host, to validate the behavior when the host is not set.
// This empty context can be used setting the currentContextIndex to a number greater than contextsCount.
func setupDockerContexts(t *testing.T, currentContextIndex int, contextsCount int) {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir) // Windows support

	configDir, err := config.Dir()
	require.NoError(t, err)

	tempMkdirAll(t, configDir)

	configJSON := filepath.Join(configDir, config.FileName)

	const baseContext = "context"

	// default config.json with no current context
	configBytes := `{"currentContext": ""}`

	if currentContextIndex <= contextsCount {
		configBytes = fmt.Sprintf(`{
	"currentContext": "%s%d"
}`, baseContext, currentContextIndex)
	}

	err = os.WriteFile(configJSON, []byte(configBytes), 0o644)
	require.NoError(t, err)

	metaDir, err := metaRoot()
	require.NoError(t, err)

	tempMkdirAll(t, metaDir)

	// first index is 1
	for i := 1; i <= contextsCount; i++ {
		createDockerContext(t, metaDir, baseContext, i, fmt.Sprintf("tcp://127.0.0.1:%d", i))
	}

	// add a context with no host
	createDockerContext(t, metaDir, baseContext, contextsCount+1, "")
}

// createDockerContext creates a Docker context with the specified name and host
func createDockerContext(t *testing.T, metaDir, baseContext string, index int, host string) {
	t.Helper()

	contextDir := filepath.Join(metaDir, fmt.Sprintf("context%d", index))
	tempMkdirAll(t, contextDir)

	context := fmt.Sprintf(`{"Name":"%s%d","Metadata":{"Description":"Testcontainers Go %d"},"Endpoints":{"docker":{"Host":"%s","SkipTLSVerify":false}}}`,
		baseContext, index, index, host)
	err := os.WriteFile(filepath.Join(contextDir, "meta.json"), []byte(context), 0o644)
	require.NoError(t, err)
}

func tempMkdirAll(t *testing.T, dir string) {
	t.Helper()

	err := os.MkdirAll(dir, 0o755)
	require.NoError(t, err)
}
