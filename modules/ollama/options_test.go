package ollama_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/ollama"
)

func TestWithUseLocal(t *testing.T) {
	req := testcontainers.GenericContainerRequest{}

	t.Run("empty", func(t *testing.T) {
		opt := ollama.WithUseLocal(nil)
		err := opt.Customize(&req)
		require.NoError(t, err)
		require.Empty(t, req.Env)
	})

	t.Run("valid", func(t *testing.T) {
		opt := ollama.WithUseLocal(map[string]string{"OLLAMA_MODELS": "/path/to/models", "OLLAMA_HOST": "localhost:1234"})
		err := opt.Customize(&req)
		require.NoError(t, err)
		require.Equal(t, "/path/to/models", req.Env["OLLAMA_MODELS"])
		require.Equal(t, "localhost:1234", req.Env["OLLAMA_HOST"])
	})
}
