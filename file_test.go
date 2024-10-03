// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"archive/tar"
	_ "embed"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_IsDir(t *testing.T) {
	type cases struct {
		filepath string
		expected bool
		err      error
	}

	tests := []cases{
		{
			filepath: "testdata",
			expected: true,
			err:      nil,
		},
		{
			filepath: "docker.go",
			expected: false,
			err:      nil,
		},
		{
			filepath: "foobar.doc",
			expected: false,
			err:      fmt.Errorf("does not exist"),
		},
	}

	for _, test := range tests {
		t.Run(test.filepath, func(t *testing.T) {
			result, err := isDir(test.filepath)
			if test.err != nil {
				require.Error(t, err, "expected error")
			} else {
				require.NoError(t, err, "not expected error")
			}
			assert.Equal(t, test.expected, result)
		})
	}
}

//go:embed testdata/Dockerfile
var dockerFile []byte

func Test_TarFile(t *testing.T) {
	// Mock modTime so we can create a sample test.
	now := time.Unix(time.Now().Unix(), 0)
	oldModTime := modTime
	modTime = func() time.Time {
		return now
	}
	t.Cleanup(func() {
		modTime = oldModTime
	})

	filename := "Docker.file"
	tarReader, err := tarFile(filename, dockerFile, 0o755)
	require.NoError(t, err)

	// Decode and validate the output matches our input.
	tr := tar.NewReader(tarReader)
	header, err := tr.Next()
	require.NoError(t, err)
	expectedHeader := &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     filename,
		Size:     int64(len(dockerFile)),
		ModTime:  now,
		Format:   tar.FormatUSTAR,
		Mode:     0o755,
	}
	require.Equal(t, expectedHeader, header)
	data, err := io.ReadAll(tr)
	require.NoError(t, err)
	require.Equal(t, dockerFile, data)
}
