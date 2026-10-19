// This test is testing very internal logic that should not be exported away from this package. We'll
// leave it in the main testcontainers package. Do not use for user facing examples.
package testcontainers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
			err:      errors.New("does not exist"),
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

func Test_TarDir(t *testing.T) {
	originalSrc := filepath.Join(".", "testdata")
	tests := []struct {
		abs bool
	}{
		{
			abs: false,
		},
		{
			abs: true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("TarDir with abs=%t", test.abs), func(t *testing.T) {
			src := originalSrc
			if test.abs {
				absSrc, err := filepath.Abs(src)
				require.NoError(t, err)

				src = absSrc
			}

			buff, err := tarDir(src, 0o755)
			require.NoError(t, err)

			tmpDir := filepath.Join(t.TempDir(), "subfolder")
			err = untar(tmpDir, bytes.NewReader(buff.Bytes()))
			require.NoError(t, err)

			srcFiles, err := os.ReadDir(src)
			require.NoError(t, err)

			for _, srcFile := range srcFiles {
				if srcFile.IsDir() {
					continue
				}
				srcBytes, err := os.ReadFile(filepath.Join(src, srcFile.Name()))
				require.NoError(t, err)

				untarBytes, err := os.ReadFile(filepath.Join(tmpDir, "testdata", srcFile.Name()))
				require.NoError(t, err)
				assert.Equal(t, srcBytes, untarBytes)
			}
		})
	}
}

func Test_TarFile(t *testing.T) {
	t.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tarinsecurepath=0")

	b, err := os.ReadFile(filepath.Join(".", "testdata", "Dockerfile"))
	require.NoError(t, err)

	modTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name     string
		path     string
		wantName string
		wantErr  string
	}{
		{name: "relative", path: "Docker.file", wantName: "Docker.file"},
		{name: "relative nested", path: "test/Docker.file", wantName: "test/Docker.file"},
		{name: "absolute nested", path: "/test/Docker.file", wantName: "test/Docker.file"},
		{name: "clean absolute", path: "/test/../tmp/./Docker.file", wantName: "tmp/Docker.file"},
		{name: "clean relative", path: "test/../tmp/./Docker.file", wantName: "tmp/Docker.file"},
		{name: "empty", path: "", wantErr: "invalid container file path"},
		{name: "dot", path: ".", wantErr: "invalid container file path"},
		{name: "clean dot", path: "test/..", wantErr: "invalid container file path"},
		{name: "root", path: "/", wantErr: "invalid container file path"},
		{name: "parent", path: "..", wantErr: "escapes the archive root"},
		{name: "parent prefix", path: "../Docker.file", wantErr: "escapes the archive root"},
		{name: "clean traversal", path: "test/../../Docker.file", wantErr: "escapes the archive root"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buff, err := tarFile(test.path, func(tw io.Writer) error {
				_, err := tw.Write(b)
				return err
			}, int64(len(b)), 0o755, modTime)
			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}
			require.NoError(t, err)

			gzr, err := gzip.NewReader(bytes.NewReader(buff.Bytes()))
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, gzr.Close()) })
			tr := tar.NewReader(gzr)

			header, err := tr.Next()
			require.NoError(t, err)
			require.Equal(t, test.wantName, header.Name)
			require.Equal(t, int64(len(b)), header.Size)
			require.Equal(t, int64(0o755), header.Mode)
			require.True(t, modTime.Equal(header.ModTime), "got %s", header.ModTime)
			got, err := io.ReadAll(tr)
			require.NoError(t, err)
			require.Equal(t, b, got)
			_, err = tr.Next()
			require.ErrorIs(t, err, io.EOF)
		})
	}
}

// untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func untar(dst string, r io.Reader) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if it's a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0o755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; deferring would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
