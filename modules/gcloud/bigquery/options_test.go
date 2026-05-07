package bigquery

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

func TestWithDataYAML(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{},
		}

		err := WithDataYAML(bytes.NewReader([]byte("")))(req)
		require.NoError(t, err)
		require.Contains(t, req.Cmd, "--data-from-yaml")
		require.Len(t, req.Files, 1)
	})

	t.Run("double-calls-errors", func(t *testing.T) {
		req := &testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{},
		}

		err := WithDataYAML(bytes.NewReader([]byte("")))(req)
		require.NoError(t, err)
		require.Contains(t, req.Cmd, "--data-from-yaml")
		require.Len(t, req.Files, 1)

		err = WithDataYAML(bytes.NewReader([]byte("")))(req)
		require.Error(t, err)
		require.Len(t, req.Files, 1)
	})
}
