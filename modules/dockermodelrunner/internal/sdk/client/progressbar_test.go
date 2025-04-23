package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithProgressBar(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		w := bytes.NewBuffer(nil)
		we := bytes.NewBuffer(nil)
		err := WithProgressBar(w, we, 100)(&pullOptions{})
		require.NoError(t, err)
	})

	t.Run("error/writer-nil", func(t *testing.T) {
		we := bytes.NewBuffer(nil)
		err := WithProgressBar(nil, we, 100)(&pullOptions{})
		require.Error(t, err)
	})

	t.Run("error/error-writer-nil", func(t *testing.T) {
		w := bytes.NewBuffer(nil)
		err := WithProgressBar(w, nil, 100)(&pullOptions{})
		require.Error(t, err)
	})
}
