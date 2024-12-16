package ollama

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun_localWithCustomLogFileError(t *testing.T) {
	t.Run("terminate/close-log-error", func(t *testing.T) {
		// Create a temporary file for testing
		f, err := os.CreateTemp(t.TempDir(), "test-log-*")
		require.NoError(t, err)

		// Close the file before termination to force a "file already closed" error
		err = f.Close()
		require.NoError(t, err)

		c := &OllamaContainer{
			localCtx: &localContext{
				logFile: f,
			},
		}
		err = c.Terminate(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "close log:")
	})

	t.Run("terminate/log-file-not-removable", func(t *testing.T) {
		// Create a temporary file for testing
		f, err := os.CreateTemp(t.TempDir(), "test-log-*")
		require.NoError(t, err)
		defer func() {
			// Cleanup: restore permissions
			os.Chmod(filepath.Dir(f.Name()), 0700)
		}()

		// Make the file read-only and its parent directory read-only
		// This should cause removal to fail on most systems
		dir := filepath.Dir(f.Name())
		require.NoError(t, os.Chmod(dir, 0500))

		c := &OllamaContainer{
			localCtx: &localContext{
				logFile: f,
			},
		}
		err = c.Terminate(context.Background())
		require.Error(t, err)
		require.ErrorContains(t, err, "remove log:")
	})
}
