package testcontainers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_IsDir(t *testing.T) {
	type cases struct {
		filepath string
		expected bool
		err      error
	}

	tests := []cases{
		{
			filepath: "testresources",
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
				assert.NotNil(t, err, "expected error")
			} else {
				assert.Nil(t, err, "not expected error")
			}
			assert.Equal(t, test.expected, result)
		})
	}
}

func Test_TarDir(t *testing.T) {
	src := filepath.Join(".", "testresources")

	buff, err := tarDir(src, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := filepath.Join(t.TempDir(), "subfolder")
	err = untar(tmpDir, bytes.NewReader(buff.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	srcFiles, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}

	for _, srcFile := range srcFiles {
		srcBytes, err := ioutil.ReadFile(filepath.Join(src, srcFile.Name()))
		if err != nil {
			t.Fatal(err)
		}

		untarBytes, err := ioutil.ReadFile(filepath.Join(tmpDir, "testresources", srcFile.Name()))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, srcBytes, untarBytes)
	}
}

func Test_TarFile(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join(".", "testresources", "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}

	buff, err := tarFile(b, "Docker.file", 0755)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	err = untar(tmpDir, bytes.NewReader(buff.Bytes()))
	if err != nil {
		t.Fatal(err)
	}

	untarBytes, err := ioutil.ReadFile(filepath.Join(tmpDir, "Docker.file"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, b, untarBytes)
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

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
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

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
