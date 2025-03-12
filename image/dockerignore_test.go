package image

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDockerIgnore(t *testing.T) {
	assertions := func(t *testing.T, filePath string, exists bool, expectedErr error, expectedExcluded []string) {
		t.Helper()

		ok, excluded, err := ParseDockerIgnore(filePath)
		require.Equal(t, exists, ok)
		require.Equal(t, expectedErr, err)
		require.Equal(t, expectedExcluded, excluded)
	}

	t.Run("file-exists", func(t *testing.T) {
		assertions(t, "./testdata/dockerignore", true, nil, []string{"vendor", "foo", "bar"})
	})

	t.Run("file-exists-including-comments", func(t *testing.T) {
		assertions(t, "./testdata", true, nil, []string{"Dockerfile", "echo.Dockerfile"})
	})

	t.Run("file-does-not-exists", func(t *testing.T) {
		assertions(t, "./testdata/data", false, nil, nil)
	})
}
