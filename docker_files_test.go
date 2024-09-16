package testcontainers_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCopyFileToContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// copyFileOnCreate {
	r, err := os.Open("testdata/hello.sh")
	require.NoError(t, err)

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					Reader:            r,
					ContainerFilePath: "/hello.sh",
					FileMode:          0o700,
				},
			},
			Cmd:        []string{"bash", "/hello.sh"},
			WaitingFor: wait.ForLog("done"),
		},
		Started: true,
	})
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestCopyFileToRunningContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// copyFileAfterCreate {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash:5.2.26",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "testdata/waitForHello.sh",
					ContainerFilePath: "/waitForHello.sh",
					FileMode:          0o700,
				},
			},
			Cmd: []string{"bash", "/waitForHello.sh"},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	err = ctr.CopyFileToContainer(ctx, "testdata/hello.sh", "/scripts/hello.sh", 0o700)
	// }
	require.NoError(t, err)

	waitForDone(ctx, t, ctr)
}

func TestCopyDirectoryToContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// copyDirectoryToContainer {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath: "testdata",
					// ContainerFile cannot create the parent directory, so we copy the scripts
					// to the root of the container instead. Make sure to create the container directory
					// before you copy a host directory on create.
					ContainerFilePath: "/",
					FileMode:          0o700,
				},
			},
			Cmd:        []string{"bash", "/testdata/hello.sh"},
			WaitingFor: wait.ForLog("done"),
		},
		Started: true,
	})
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	waitForDone(ctx, t, ctr)
}

func TestCopyDirectoryToRunningContainerAsFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// copyDirectoryToRunningContainerAsFile {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "testdata/waitForHello.sh",
					ContainerFilePath: "/waitForHello.sh",
					FileMode:          0o700,
				},
			},
			Cmd: []string{"bash", "/waitForHello.sh"},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// Because the container path is a directory, it will use the copy dir method as fallback.
	err = ctr.CopyFileToContainer(ctx, "testdata", "/scripts", 0o700)
	require.NoError(t, err)
	// }

	waitForDone(ctx, t, ctr)
}

func TestCopyDirectoryToRunningContainerAsDir(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// copyDirectoryToRunningContainerAsDir {
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      "testdata/waitForHello.sh",
					ContainerFilePath: "/waitForHello.sh",
					FileMode:          0o700,
				},
			},
			Cmd: []string{"bash", "/waitForHello.sh"},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	err = ctr.CopyDirToContainer(ctx, "testdata", "/scripts", 0o700)
	require.NoError(t, err)
	// }

	waitForDone(ctx, t, ctr)
}

func TestCopyHostPathTo(t *testing.T) {
	ctx := context.Background()
	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	require.NoError(t, err)

	t.Run("dir-to-dir", func(t *testing.T) {
		// This should copy the contents of testdata to /scripts in the container.
		ctr := createWaitForHelloContainer(t)
		err := ctr.CopyHostPathTo(ctx, "testdata", "/scripts")
		require.NoError(t, err)
		waitForDone(ctx, t, ctr)
	})

	t.Run("file-to-non-existent-dest", func(t *testing.T) {
		// This should copy testdata/hello.sh to the file named /scripts in the container as it does not exist.
		ctr := createWaitForHelloContainer(t)
		err := ctr.CopyHostPathTo(ctx, "testdata/hello.sh", "/scripts")
		require.NoError(t, err)

		stat, err := client.ContainerStatPath(ctx, ctr.GetContainerID(), "/scripts")
		require.NoError(t, err)
		require.True(t, stat.Mode.IsRegular())
		stat, err = client.ContainerStatPath(ctx, ctr.GetContainerID(), "/scripts/hello.sh")
		require.Error(t, err)
	})

	t.Run("file-to-non-existent-dir", func(t *testing.T) {
		// This should assert that /scripts/ is a directory and try copy testdata/hello.sh to /scripts/hello.sh
		// failing as /scripts/ does not exist.
		ctr := createWaitForHelloContainer(t)
		err := ctr.CopyHostPathTo(ctx, "testdata/hello.sh", "/scripts/")
		require.ErrorIs(t, err, archive.ErrDirNotExists)
	})

	t.Run("file-to-file-dir-not-found", func(t *testing.T) {
		// This should fail as /scripts does not exist.
		ctr := createWaitForHelloContainer(t)
		err := ctr.CopyHostPathTo(ctx, "testdata/hello.sh", "/scripts/hello.sh")
		require.Error(t, err)
		require.True(t, errdefs.IsNotFound(err))
	})

	t.Run("file-to-file", func(t *testing.T) {
		// Creating the required directory first, should copy the file to the correct location.
		ctr := createWaitForHelloContainer(t)

		// Create the directory first.
		_, _, err := ctr.Exec(ctx, []string{"mkdir", "/scripts"})
		require.NoError(t, err)

		err = ctr.CopyHostPathTo(ctx, "testdata/hello.sh", "/scripts/hello.sh")
		require.NoError(t, err)
		waitForDone(ctx, t, ctr)
	})
}

// createWaitForHelloContainer creates a container that waits for the file /scripts/hello.sh.
func createWaitForHelloContainer(t *testing.T) testcontainers.Container {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{{
				HostFilePath:      "testdata/waitForHello.sh",
				ContainerFilePath: "/waitForHello.sh",
				FileMode:          0o700,
			}},
			Cmd: []string{"bash", "/waitForHello.sh"},
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	return ctr
}

// waitForDone waits for the container to output "done" into its log, which
// indicates the file was correctly created and then run.
func waitForDone(ctx context.Context, t *testing.T, ctr testcontainers.Container) {
	t.Helper()

	err := wait.ForLog("done").WithStartupTimeout(2*time.Second).WaitUntilReady(ctx, ctr)
	require.NoError(t, err)
}
