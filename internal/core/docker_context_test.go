package core

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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
func setupDockerContexts(t *testing.T, currentContextIndex int, contextsCount int) {
	t.Helper()

	configDir := filepath.Join(getHomeDir(), configFileDir)

	err := createTmpDir(configDir)
	require.NoError(t, err)

	configJson := filepath.Join(configDir, "config.json")

	const baseContext = "context"

	configBytes := fmt.Sprintf(`{
	"currentContext": "%s%d"
}`, baseContext, currentContextIndex)

	err = os.WriteFile(configJson, []byte(configBytes), 0o644)
	require.NoError(t, err)

	metaDir := filepath.Join(configDir, contextsDir, metadataDir)

	err = createTmpDir(metaDir)
	require.NoError(t, err)

	// first index is 1
	for i := 1; i <= contextsCount; i++ {
		contextDir := filepath.Join(metaDir, fmt.Sprintf("context%d", i))
		err = createTmpDir(contextDir)
		require.NoError(t, err)

		context := fmt.Sprintf(`{"Name":"%s%d","Metadata":{"Description":"Testcontainers Go %d"},"Endpoints":{"docker":{"Host":"tcp://127.0.0.1:%d","SkipTLSVerify":false}}}`, baseContext, i, i, i)
		err = os.WriteFile(filepath.Join(contextDir, "meta.json"), []byte(context), 0o644)
		require.NoError(t, err)
	}
}
