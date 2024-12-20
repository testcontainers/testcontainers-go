package ollama_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestWithUseLocal(t *testing.T) {
	req := testcontainers.GenericContainerRequest{}

	t.Run("keyVal/valid", func(t *testing.T) {
		opt := ollama.WithUseLocal("OLLAMA_MODELS=/path/to/models")
		err := opt.Customize(&req)
		require.NoError(t, err)
		require.Equal(t, "/path/to/models", req.Env["OLLAMA_MODELS"])
	})

	t.Run("keyVal/invalid", func(t *testing.T) {
		opt := ollama.WithUseLocal("OLLAMA_MODELS")
		err := opt.Customize(&req)
		require.Error(t, err)
	})

	t.Run("keyVal/valid/multiple", func(t *testing.T) {
		opt := ollama.WithUseLocal("OLLAMA_MODELS=/path/to/models", "OLLAMA_HOST=localhost")
		err := opt.Customize(&req)
		require.NoError(t, err)
		require.Equal(t, "/path/to/models", req.Env["OLLAMA_MODELS"])
		require.Equal(t, "localhost", req.Env["OLLAMA_HOST"])
	})

	t.Run("keyVal/valid/multiple-equals", func(t *testing.T) {
		opt := ollama.WithUseLocal("OLLAMA_MODELS=/path/to/models", "OLLAMA_HOST=localhost=127.0.0.1")
		err := opt.Customize(&req)
		require.NoError(t, err)
		require.Equal(t, "/path/to/models", req.Env["OLLAMA_MODELS"])
		require.Equal(t, "localhost=127.0.0.1", req.Env["OLLAMA_HOST"])
	})

	t.Run("keyVal/invalid/multiple", func(t *testing.T) {
		opt := ollama.WithUseLocal("OLLAMA_MODELS=/path/to/models", "OLLAMA_HOST")
		err := opt.Customize(&req)
		require.Error(t, err)
	})
}
