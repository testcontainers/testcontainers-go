//go:build !windows

// TODO: remove this file when [Container.CopyDirToContainer] and
// [Container.CopyFileToContainer] have been removed.

package testcontainers

import (
	"path/filepath"
)

func normalizePath(path string) string {
	return filepath.ToSlash(path)
}
