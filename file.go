package testcontainers

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// isDir returns true if the path is a directory and false otherwise.
func isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("stat: %w", err)
	}

	return fileInfo.IsDir(), nil
}

// Ensure fileInfo implements fs.FileInfo.
var _ fs.FileInfo = &fileInfo{}

// modTime is a variable so we can mock it in tests.
var modTime = time.Now

// fileInfo implements fs.FileInfo with fixed details.
type fileInfo struct {
	name     string
	fileMode fs.FileMode
	modTime  time.Time
	size     int64
}

// newFileInfo creates a new fileInfo with the given details.
func newFileInfo(name string, size int64, fileMode fs.FileMode) *fileInfo {
	return &fileInfo{
		name:     filepath.Base(name),
		size:     size,
		fileMode: fileMode,
		modTime:  modTime(),
	}
}

func (f *fileInfo) Size() int64        { return f.size }
func (f *fileInfo) Name() string       { return f.name }
func (f *fileInfo) IsDir() bool        { return false }
func (f *fileInfo) Mode() fs.FileMode  { return f.fileMode }
func (f *fileInfo) ModTime() time.Time { return f.modTime }
func (f *fileInfo) Sys() any           { return nil }

// tarFile creates a tar archive containing a single file with the given
// details and returns a reader for it.
func tarFile(filePath string, fileContent []byte, fileMode fs.FileMode) (io.ReadCloser, error) {
	fi := newFileInfo(filePath, int64(len(fileContent)), fileMode)
	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return nil, fmt.Errorf("file info header: %w", err)
	}

	// Set the name to the fully qualified so we have the right path in the tar.
	hdr.Name = filePath

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

	if _, err := tw.Write(fileContent); err != nil {
		return nil, fmt.Errorf("write file content: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("close tar: %w", err)
	}

	return io.NopCloser(&buf), nil
}
