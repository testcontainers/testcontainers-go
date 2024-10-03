package testcontainers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/system"
)

// CopyHostPathTo copies the contents of a hostPath to containerPath in the container
// with the given options.
// If the parent of the containerPath does not exist an error is returned.
func (c *DockerContainer) CopyHostPathTo(ctx context.Context, hostPath, containerPath string, options ...CopyToOption) error {
	srcPath, err := resolveLocalPath(hostPath)
	if err != nil {
		return err
	}

	copyOptions, dstInfo, err := c.copyToDetails(ctx, containerPath, options...)
	if err != nil {
		return err
	}

	// Prepare source copy info.
	srcInfo, err := archive.CopyInfoSourcePath(srcPath, copyOptions.followLink)
	if err != nil {
		return fmt.Errorf("copy info source path: %w", err)
	}

	// TODO: Do we want to support copying to a directory that does not exist?
	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return fmt.Errorf("tar resource: %w", err)
	}
	defer srcArchive.Close()

	// With the stat info about the local source as well as the
	// destination, we have enough information to know whether we need to
	// alter the archive that we upload so that when the server extracts
	// it to the specified directory in the container we get the desired
	// copy behaviour.

	// See comments in the implementation of [archive.PrepareArchiveCopy]
	// for exactly what goes into deciding how and whether the source
	// archive needs to be altered for the correct copy behaviour when it is
	// extracted. This function also infers from the source and destination
	// info which directory to extract to, which may be the parent of the
	// destination that the user specified.
	// TODO: replace prepareArchiveCopyFilePerm with [archive.PrepareArchiveCopy].
	dstDir, preparedArchive, err := prepareArchiveCopyFilePerm(srcArchive, srcInfo, *dstInfo, copyOptions.fileMode)
	if err != nil {
		return fmt.Errorf("prepare archive copy: %w", err)
	}
	defer preparedArchive.Close()

	return c.copyTo(ctx, preparedArchive, dstDir, copyOptions)
}

// copyTarTo copies the contents of a tar stream to a destination in the container.
func (c *DockerContainer) copyTarTo(ctx context.Context, tar io.ReadCloser, targetDir string, options ...CopyToOption) error {
	copyOptions, dstInfo, err := c.copyToDetails(ctx, targetDir, options...)
	if err != nil {
		return err
	}

	if !dstInfo.IsDir {
		return fmt.Errorf("destination %q must be a directory", targetDir)
	}

	return c.copyTo(ctx, tar, dstInfo.Path, copyOptions)
}

// resolveLocalPath resolves the absolute path of a localPath and preserves the trailing dot or separator.
// See [archive.PreserveTrailingDotOrSeparator] for more information.
func resolveLocalPath(localPath string) (absPath string, err error) {
	if absPath, err = filepath.Abs(localPath); err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	return archive.PreserveTrailingDotOrSeparator(absPath, localPath), nil
}

// validateOutputPath validates the output paths file mode.
func validateOutputPath(fileMode os.FileMode) error {
	switch {
	case fileMode&os.ModeDevice != 0:
		return errors.New("got a device")
	case fileMode&os.ModeIrregular != 0:
		return errors.New("got an irregular file")
	}
	return nil
}

// copyToDetails returns the destination info and copying options.
func (c *DockerContainer) copyToDetails(ctx context.Context, dstPath string, options ...CopyToOption) (copyOptions *copyToOptions, destInfo *archive.CopyInfo, err error) {
	// Prepare destination copy info by stat-ing the container path.
	destInfo = &archive.CopyInfo{Path: dstPath}
	dstStat, err := c.provider.client.ContainerStatPath(ctx, c.ID, dstPath)
	defer c.provider.Close()
	if err != nil {
		// TODO: Validate that this is the correct error to ignore.
		if !errdefs.IsNotFound(err) {
			return nil, nil, fmt.Errorf("container stat path: %w", err)
		}

		// Ignore any error and assume that the parent directory of the destination
		// path exists, in which case the copy may still succeed. If there is any
		// type of conflict (e.g., non-directory overwriting an existing directory
		// or vice versa) the extraction will fail. If the destination simply did
		// not exist, but the parent directory does, the extraction will still
		// succeed.
	} else {
		if dstStat.Mode&os.ModeSymlink != 0 {
			// The destination is a symbolic link so evaluate it.
			linkTarget := dstStat.LinkTarget
			if !system.IsAbs(linkTarget) {
				// Join with the parent directory.
				dstParent, _ := archive.SplitPathDirEntry(dstPath)
				linkTarget = filepath.Join(dstParent, linkTarget)
			}

			destInfo.Path = linkTarget
			dstStat, err = c.provider.client.ContainerStatPath(ctx, c.ID, linkTarget)
		}

		destInfo.Exists, destInfo.IsDir = true, dstStat.Mode.IsDir()
	}

	// Validate the destination path.
	if err := validateOutputPath(dstStat.Mode); err != nil {
		return nil, nil, fmt.Errorf("destination %q must be a directory or a regular file: %w", dstPath, err)
	}

	copyOptions = &copyToOptions{}
	for _, option := range options {
		option(copyOptions)
	}

	return copyOptions, destInfo, nil
}

// copyTo copies content to destination in the container.
func (c *DockerContainer) copyTo(ctx context.Context, content io.ReadCloser, path string, copyOptions *copyToOptions) error {
	var opts container.CopyToContainerOptions
	opts.CopyUIDGID = copyOptions.copyUIDGID
	opts.AllowOverwriteDirWithFile = copyOptions.allowOverwriteDirWithFile

	if err := c.provider.client.CopyToContainer(ctx, c.ID, path, content, opts); err != nil {
		return fmt.Errorf("copy to container: %w", err)
	}

	return nil
}
