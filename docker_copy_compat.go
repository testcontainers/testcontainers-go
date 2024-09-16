// TODO: remove this file when [Container.CopyDirToContainer] and
// [Container.CopyFileToContainer] have been removed.
//
// This provides [prepareArchiveCopyFilePerm] which replicates
// [archive.PrepareArchiveCopy] adding support for updating the file
// permission.

package testcontainers

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/archive"
)

const (
	// modeSafe is all fs.FileMode bits which are safe to update.
	modeSafe = fs.ModePerm | fs.ModeSetgid | fs.ModeSetuid | fs.ModeSticky

	// tarModeSafe is all file mode bits which are safe to update.
	tarModeSafe = int64(fs.ModePerm | tarSetgid | tarSetuid | tarSticky)

	// Mode constants from the USTAR spec:
	// See http://pubs.opengroup.org/onlinepubs/9699919799/utilities/pax.html#tag_20_92_13_06
	// tarSetuid is USTAR set uid
	tarSetuid = 0o4000
	// tarSetgid is USTAR set gid
	tarSetgid = 0o2000
	// tarSticky is USTAR sticky
	tarSticky = 0o1000
)

// validateFileMode validates that only valid bits are set in fileMode.
func validateFileMode(fileMode int64) error {
	if otherBits := fs.FileMode(fileMode) & ^modeSafe; otherBits != 0 {
		return fmt.Errorf("invalid file mode %q: unexpected %q", fs.FileMode(fileMode), fs.FileMode(otherBits))
	}

	return nil
}

// prepareArchiveCopyFilePerm prepares the given srcContent archive, which should
// contain the archived resource described by srcInfo, to the destination
// described by dstInfo. Returns the possibly modified content archive along
// with the path to the destination directory which it should be extracted to.
// It replicates the functionality of [archive.PrepareArchiveCopy] adding the
// optional fileMode parameter allowing the Permissions of files added to be
// updated.
func prepareArchiveCopyFilePerm(srcContent io.Reader, srcInfo, dstInfo archive.CopyInfo, fileMode fs.FileMode) (dstDir string, content io.ReadCloser, err error) {
	// Ensure in platform semantics
	srcInfo.Path = normalizePath(srcInfo.Path)
	dstInfo.Path = normalizePath(dstInfo.Path)

	// Separate the destination path between its directory and base
	// components in case the source archive contents need to be rebased.
	dstDir, dstBase := archive.SplitPathDirEntry(dstInfo.Path)
	_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)

	switch {
	case dstInfo.Exists && dstInfo.IsDir:
		// The destination exists as a directory. No alteration
		// to srcContent is needed as its contents can be
		// simply extracted to the destination directory.
		if fileMode != 0 {
			return dstInfo.Path, updateAchieveFileMode(srcContent, fileMode), nil
		}
		// TODO: fix me
		return dstInfo.Path, io.NopCloser(srcContent), nil
	case dstInfo.Exists && srcInfo.IsDir:
		// The destination exists as some type of file and the source
		// content is a directory. This is an error condition since
		// you cannot copy a directory to an existing file location.
		return "", nil, archive.ErrCannotCopyDir
	case dstInfo.Exists:
		// The destination exists as some type of file and the source content
		// is also a file. The source content entry will have to be renamed to
		// have a basename which matches the destination path's basename.
		if len(srcInfo.RebaseName) != 0 {
			srcBase = srcInfo.RebaseName
		}
		return dstDir, rebaseArchiveEntries(srcContent, srcBase, dstBase, fileMode), nil
	case srcInfo.IsDir:
		// The destination does not exist and the source content is an archive
		// of a directory. The archive should be extracted to the parent of
		// the destination path instead, and when it is, the directory that is
		// created as a result should take the name of the destination path.
		// The source content entries will have to be renamed to have a
		// basename which matches the destination path's basename.
		if len(srcInfo.RebaseName) != 0 {
			srcBase = srcInfo.RebaseName
		}
		return dstDir, rebaseArchiveEntries(srcContent, srcBase, dstBase, fileMode), nil
	case assertsDirectory(dstInfo.Path):
		// The destination does not exist and is asserted to be created as a
		// directory, but the source content is not a directory. This is an
		// error condition since you cannot create a directory from a file
		// source.
		return "", nil, archive.ErrDirNotExists
	default:
		// The last remaining case is when the destination does not exist, is
		// not asserted to be a directory, and the source content is not an
		// archive of a directory. It this case, the destination file will need
		// to be created when the archive is extracted and the source content
		// entry will have to be renamed to have a basename which matches the
		// destination path's basename.
		if len(srcInfo.RebaseName) != 0 {
			srcBase = srcInfo.RebaseName
		}
		return dstDir, rebaseArchiveEntries(srcContent, srcBase, dstBase, fileMode), nil
	}
}

// rebaseArchiveEntries rewrites the given srcContent archive replacing
// an occurrence of oldBase with newBase at the beginning of entry names
// and setting file Mode bits to fileMode.
func rebaseArchiveEntries(srcContent io.Reader, oldBase, newBase string, fileMode fs.FileMode) io.ReadCloser {
	if oldBase == string(os.PathSeparator) {
		// If oldBase specifies the root directory, use an empty string as
		// oldBase instead so that newBase doesn't replace the path separator
		// that all paths will start with.
		oldBase = ""
	}

	updated, w := io.Pipe()

	go func() {
		tarReader := tar.NewReader(srcContent)
		tarWriter := tar.NewWriter(w)

		for {
			hdr, err := tarReader.Next()
			if errors.Is(err, io.EOF) {
				// Signals end of archive.
				tarWriter.Close()
				w.Close()
				return
			}
			if err != nil {
				w.CloseWithError(err)
				return
			}

			// srcContent tar stream, as served by TarWithOptions(), is
			// definitely in PAX format, but tar.Next() mistakenly guesses it
			// as USTAR, which creates a problem: if the newBase is >100
			// characters long, WriteHeader() returns an error like
			// "archive/tar: cannot encode header: Format specifies USTAR; and USTAR cannot encode Name=...".
			//
			// To fix, set the format to PAX here. See docker/for-linux issue #484.
			hdr.Format = tar.FormatPAX
			hdr.Name = strings.Replace(hdr.Name, oldBase, newBase, 1)
			if fileMode > 0 {
				setFileMode(hdr, fileMode)
			}
			if hdr.Typeflag == tar.TypeLink {
				hdr.Linkname = strings.Replace(hdr.Linkname, oldBase, newBase, 1)
			}

			if err = tarWriter.WriteHeader(hdr); err != nil {
				w.CloseWithError(err)
				return
			}

			if _, err = io.Copy(tarWriter, tarReader); err != nil {
				w.CloseWithError(err)
				return
			}
		}
	}()

	return updated
}

// setFileMode sets the hdr.Mode from the fileMode, converting
// from golang to USTAR semantics.
func setFileMode(hdr *tar.Header, fileMode fs.FileMode) {
	tarMode := fileMode.Perm()
	if fileMode&fs.ModeSetgid != 0 {
		tarMode |= tarSetgid
	}

	if fileMode&fs.ModeSetuid != 0 {
		tarMode |= tarSetuid
	}

	if fileMode&fs.ModeSticky != 0 {
		tarMode |= tarSticky
	}

	hdr.Mode &= ^tarModeSafe
	hdr.Mode |= int64(tarMode)
}

// copyToFileMode instructs the copy operation to override the file permissions.
func copyToFileMode(fileMode int64) CopyToOption {
	return func(o *copyToOptions) {
		// Ensure that fileMode only includes bits which can safely be updated.
		o.fileMode = fs.FileMode(fileMode) & modeSafe
	}
}

// assertsDirectory returns whether the given path is
// asserted to be a directory, i.e., the path ends with
// a trailing '/' or `/.`, assuming a path separator of `/`.
func assertsDirectory(path string) bool {
	return hasTrailingPathSeparator(path) || specifiesCurrentDir(path)
}

// hasTrailingPathSeparator returns whether the given
// path ends with the system's path separator character.
func hasTrailingPathSeparator(path string) bool {
	return len(path) > 0 && path[len(path)-1] == filepath.Separator
}

// specifiesCurrentDir returns whether the given path specifies
// a "current directory", i.e., the last path segment is `.`.
func specifiesCurrentDir(path string) bool {
	return filepath.Base(path) == "."
}

// updateAchieveFileMode rewrites the given srcContent archive updating
// the file modes to be fileMode.
func updateAchieveFileMode(srcContent io.Reader, fileMode fs.FileMode) io.ReadCloser {
	updated, w := io.Pipe()

	go func() {
		tarReader := tar.NewReader(srcContent)
		tarWriter := tar.NewWriter(w)

		for {
			hdr, err := tarReader.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					tarWriter.Close()
					w.Close()
					return
				}

				w.CloseWithError(err)
				return
			}

			setFileMode(hdr, fileMode)
			if err = tarWriter.WriteHeader(hdr); err != nil {
				w.CloseWithError(err)
				return
			}

			if _, err = io.Copy(tarWriter, tarReader); err != nil {
				w.CloseWithError(err)
				return
			}
		}
	}()

	return updated
}
