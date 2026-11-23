package testcontainers_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testBashImage string = "bash:5.2.26"

func TestCopyFileToContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// copyFileOnCreate {
	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)

	r, err := os.Open(absPath)
	require.NoError(t, err)

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithFiles(testcontainers.ContainerFile{
			Reader:            r,
			HostFilePath:      absPath, // will be discarded internally
			ContainerFilePath: "/hello.sh",
			FileMode:          0o700,
		}),
		testcontainers.WithCmd("bash", "/hello.sh"),
		testcontainers.WithWaitStrategy(wait.ForLog("done")),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestCopyFileToRunningContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Not using the assertations here to avoid leaking the library into the example
	// copyFileAfterCreate {
	waitForPath, err := filepath.Abs(filepath.Join(".", "testdata", "waitForHello.sh"))
	require.NoError(t, err)
	helloPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	require.NoError(t, err)

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      waitForPath,
			ContainerFilePath: "/waitForHello.sh",
			FileMode:          0o700,
		}),
		testcontainers.WithCmd("bash", "/waitForHello.sh"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	err = ctr.CopyFileToContainer(ctx, helloPath, "/scripts/hello.sh", 0o700)
	// }

	require.NoError(t, err)

	// Give some time to the wait script to catch the hello script being created
	err = wait.ForLog("done").WithStartupTimeout(2*time.Second).WaitUntilReady(ctx, ctr)
	require.NoError(t, err)
}

func TestCopyFileToRunningContainerPreservesModTime(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	hostPath := filepath.Join(t.TempDir(), "hello.txt")
	require.NoError(t, os.WriteFile(hostPath, []byte("hello"), 0o600))
	modTime := time.Date(2024, time.March, 14, 15, 9, 26, 500_000_000, time.UTC)
	require.NoError(t, os.Chtimes(hostPath, modTime, modTime))

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithEntrypoint("tail", "-f", "/dev/null"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	err = ctr.CopyFileToContainer(ctx, hostPath, "/tmp/hello.txt", 0o644)
	require.NoError(t, err)

	code, reader, err := ctr.Exec(ctx, []string{"stat", "-c", "%y", "/tmp/hello.txt"}, exec.Multiplexed())
	require.NoError(t, err)
	require.Zero(t, code)

	out, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, modTime.Format("2006-01-02 15:04:05.000000000 -0700"), strings.TrimSpace(string(out)))
}

func TestCopyDirectoryToContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Not using the assertations here to avoid leaking the library into the example
	// copyDirectoryToContainer {
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	require.NoError(t, err)

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath: dataDirectory,
			// ContainerFile cannot create the parent directory, so we copy the scripts
			// to the root of the container instead. Make sure to create the container directory
			// before you copy a host directory on create.
			ContainerFilePath: "/",
			FileMode:          0o700,
		}),
		testcontainers.WithCmd("bash", "/testdata/hello.sh"),
		testcontainers.WithWaitStrategy(wait.ForLog("done")),
	)
	// }
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
}

func TestCopyDirectoryToRunningContainerAsFile(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// copyDirectoryToRunningContainerAsFile {
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	require.NoError(t, err)
	waitForPath, err := filepath.Abs(filepath.Join(dataDirectory, "waitForHello.sh"))
	require.NoError(t, err)

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      waitForPath,
			ContainerFilePath: "/waitForHello.sh",
			FileMode:          0o700,
		}),
		testcontainers.WithCmd("bash", "/waitForHello.sh"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// as the container is started, we can create the directory first
	_, _, err = ctr.Exec(ctx, []string{"mkdir", "-p", "/scripts"})
	require.NoError(t, err)

	// because the container path is a directory, it will use the copy dir method as fallback
	err = ctr.CopyFileToContainer(ctx, dataDirectory, "/scripts", 0o700)
	require.NoError(t, err)
	// }
}

func TestCopyDirectoryToRunningContainerAsDir(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Not using the assertations here to avoid leaking the library into the example
	// copyDirectoryToRunningContainerAsDir {
	waitForPath, err := filepath.Abs(filepath.Join(".", "testdata", "waitForHello.sh"))
	require.NoError(t, err)
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	require.NoError(t, err)

	ctr, err := testcontainers.Run(ctx, testBashImage,
		testcontainers.WithFiles(testcontainers.ContainerFile{
			HostFilePath:      waitForPath,
			ContainerFilePath: "/waitForHello.sh",
			FileMode:          0o700,
		}),
		testcontainers.WithCmd("bash", "/waitForHello.sh"),
	)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)

	// as the container is started, we can create the directory first
	_, _, err = ctr.Exec(ctx, []string{"mkdir", "-p", "/scripts"})
	require.NoError(t, err)

	err = ctr.CopyDirToContainer(ctx, dataDirectory, "/scripts", 0o700)
	require.NoError(t, err)
	// }
}
