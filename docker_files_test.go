package testcontainers_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCopyFileToContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// copyFileOnCreate {
	absPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	if err != nil {
		t.Fatal(err)
	}

	r, err := os.Open(absPath)
	if err != nil {
		t.Fatal(err)
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					Reader:            r,
					HostFilePath:      absPath, // will be discarded internally
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
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Not using the assertations here to avoid leaking the library into the example
	// copyFileAfterCreate {
	waitForPath, err := filepath.Abs(filepath.Join(".", "testdata", "waitForHello.sh"))
	if err != nil {
		t.Fatal(err)
	}
	helloPath, err := filepath.Abs(filepath.Join(".", "testdata", "hello.sh"))
	if err != nil {
		t.Fatal(err)
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash:5.2.26",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      waitForPath,
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

	err = ctr.CopyFileToContainer(ctx, helloPath, "/scripts/hello.sh", 0o700)
	// }

	require.NoError(t, err)

	// Give some time to the wait script to catch the hello script being created
	err = wait.ForLog("done").WithStartupTimeout(2*time.Second).WaitUntilReady(ctx, ctr)
	require.NoError(t, err)
}

func TestCopyDirectoryToContainer(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// Not using the assertations here to avoid leaking the library into the example
	// copyDirectoryToContainer {
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	if err != nil {
		t.Fatal(err)
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath: dataDirectory,
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
}

func TestCopyDirectoryToRunningContainerAsFile(t *testing.T) {
	ctx, cnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer cnl()

	// copyDirectoryToRunningContainerAsFile {
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	waitForPath, err := filepath.Abs(filepath.Join(dataDirectory, "waitForHello.sh"))
	if err != nil {
		t.Fatal(err)
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      waitForPath,
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
	if err != nil {
		t.Fatal(err)
	}
	dataDirectory, err := filepath.Abs(filepath.Join(".", "testdata"))
	if err != nil {
		t.Fatal(err)
	}

	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "docker.io/bash",
			Files: []testcontainers.ContainerFile{
				{
					HostFilePath:      waitForPath,
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

	// as the container is started, we can create the directory first
	_, _, err = ctr.Exec(ctx, []string{"mkdir", "-p", "/scripts"})
	require.NoError(t, err)

	err = ctr.CopyDirToContainer(ctx, dataDirectory, "/scripts", 0o700)
	require.NoError(t, err)
	// }
}
