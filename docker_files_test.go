package testcontainers_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCopyFileInTheRequest(t *testing.T) {
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

	ctr, err := testcontainers.Run(ctx, testcontainers.Request{
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
		Started:    true,
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

	ctr, err := testcontainers.Run(ctx, testcontainers.Request{
		Image: "docker.io/bash:5.2.26",
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      waitForPath,
				ContainerFilePath: "/waitForHello.sh",
				FileMode:          0o700,
			},
		},
		Cmd:     []string{"bash", "/waitForHello.sh"},
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

	ctr, err := testcontainers.Run(ctx, testcontainers.Request{
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
		Started:    true,
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

	ctr, err := testcontainers.Run(ctx, testcontainers.Request{
		Image: "docker.io/bash",
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      waitForPath,
				ContainerFilePath: "/waitForHello.sh",
				FileMode:          0o700,
			},
		},
		Cmd:     []string{"bash", "/waitForHello.sh"},
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

	ctr, err := testcontainers.Run(ctx, testcontainers.Request{
		Image: "docker.io/bash",
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      waitForPath,
				ContainerFilePath: "/waitForHello.sh",
				FileMode:          0o700,
			},
		},
		Cmd:     []string{"bash", "/waitForHello.sh"},
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

func TestDockerContainerCopyFileToContainer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		copiedFileName string
	}{
		{
			name:           "success copy",
			copiedFileName: "/hello_copy.sh",
		},
		{
			name:           "success copy with dir",
			copiedFileName: "/test/hello_copy.sh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
				Image:        nginxImage,
				ExposedPorts: []string{nginxDefaultPort},
				WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				Started:      true,
			})
			testcontainers.CleanupContainer(t, nginxC)
			require.NoError(t, err)

			err = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "hello.sh"), tc.copiedFileName, 700)
			require.NoError(t, err)

			c, _, err := nginxC.Exec(ctx, []string{"bash", tc.copiedFileName})
			require.NoError(t, err)
			require.Equal(t, 0, c, "File %s should exist, expected return code 0, got %v", tc.copiedFileName, c)
		})
	}
}

func TestDockerContainerCopyDirToContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:        nginxImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		Started:      true,
	})
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	p := filepath.Join(".", "testdata", "Dokerfile")

	err = nginxC.CopyDirToContainer(ctx, p, "/tmp/testdata/Dockerfile", 700)
	require.Error(t, err) // copying a file using the directory method will raise an error

	p = filepath.Join(".", "testdata")
	err = nginxC.CopyDirToContainer(ctx, p, "/tmp/testdata", 700)
	if err != nil {
		t.Fatal(err)
	}

	assertExtractedFiles(t, ctx, nginxC, p, "/tmp/testdata/")
}

func TestDockerCreateContainerWithFiles(t *testing.T) {
	ctx := context.Background()
	hostFileName := filepath.Join(".", "testdata", "hello.sh")
	copiedFileName := "/hello_copy.sh"
	tests := []struct {
		name   string
		files  []testcontainers.ContainerFile
		errMsg string
	}{
		{
			name: "success copy",
			files: []testcontainers.ContainerFile{
				{
					HostFilePath:      hostFileName,
					ContainerFilePath: copiedFileName,
					FileMode:          700,
				},
			},
		},
		{
			name: "host file not found",
			files: []testcontainers.ContainerFile{
				{
					HostFilePath:      hostFileName + "123",
					ContainerFilePath: copiedFileName,
					FileMode:          700,
				},
			},
			errMsg: "can't copy " +
				hostFileName + "123 to container: open " + hostFileName + "123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
				Image:        "nginx:1.17.6",
				ExposedPorts: []string{"80/tcp"},
				WaitingFor:   wait.ForListeningPort("80/tcp"),
				Files:        tc.files,
				Started:      false,
			})
			testcontainers.CleanupContainer(t, nginxC)

			if err != nil {
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				for _, f := range tc.files {
					require.NoError(t, err)

					hostFileData, err := os.ReadFile(f.HostFilePath)
					require.NoError(t, err)

					fd, err := nginxC.CopyFileFromContainer(ctx, f.ContainerFilePath)
					require.NoError(t, err)
					defer fd.Close()
					containerFileData, err := io.ReadAll(fd)
					require.NoError(t, err)

					require.Equal(t, hostFileData, containerFileData)
				}
			}
		})
	}
}

func TestDockerCreateContainerWithDirs(t *testing.T) {
	ctx := context.Background()
	hostDirName := "testdata"

	abs, err := filepath.Abs(filepath.Join(".", hostDirName))
	require.NoError(t, err)

	tests := []struct {
		name     string
		dir      testcontainers.ContainerFile
		hasError bool
	}{
		{
			name: "success copy directory with full path",
			dir: testcontainers.ContainerFile{
				HostFilePath:      abs,
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          700,
			},
			hasError: false,
		},
		{
			name: "success copy directory",
			dir: testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(".", hostDirName),
				ContainerFilePath: "/tmp/" + hostDirName, // the parent dir must exist
				FileMode:          700,
			},
			hasError: false,
		},
		{
			name: "host dir not found",
			dir: testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(".", "testdata123"), // does not exist
				ContainerFilePath: "/tmp/" + hostDirName,             // the parent dir must exist
				FileMode:          700,
			},
			hasError: true,
		},
		{
			name: "container dir not found",
			dir: testcontainers.ContainerFile{
				HostFilePath:      filepath.Join(".", hostDirName),
				ContainerFilePath: "/parent-does-not-exist/testdata123", // does not exist
				FileMode:          700,
			},
			hasError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
				Image:        "nginx:1.17.6",
				ExposedPorts: []string{"80/tcp"},
				WaitingFor:   wait.ForListeningPort("80/tcp"),
				Files:        []testcontainers.ContainerFile{tc.dir},
				Started:      false,
			})
			testcontainers.CleanupContainer(t, nginxC)

			require.Equal(t, (err != nil), tc.hasError)
			if err == nil {
				dir := tc.dir

				assertExtractedFiles(t, ctx, nginxC, dir.HostFilePath, dir.ContainerFilePath)
			}
		})
	}
}

func TestDockerContainerCopyToContainer(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		copiedFileName string
	}{
		{
			name:           "success copy",
			copiedFileName: "hello_copy.sh",
		},
		{
			name:           "success copy with dir",
			copiedFileName: "/test/hello_copy.sh",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
				Image:        nginxImage,
				ExposedPorts: []string{nginxDefaultPort},
				WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
				Started:      true,
			})
			testcontainers.CleanupContainer(t, nginxC)
			require.NoError(t, err)

			fileContent, err := os.ReadFile(filepath.Join(".", "testdata", "hello.sh"))
			if err != nil {
				t.Fatal(err)
			}
			err = nginxC.CopyToContainer(ctx, fileContent, tc.copiedFileName, 700)
			require.NoError(t, err)

			c, _, err := nginxC.Exec(ctx, []string{"bash", tc.copiedFileName})
			require.NoError(t, err)
			require.Equal(t, 0, c, "File %s should exist, expected return code 0, got %v", tc.copiedFileName, c)
		})
	}
}

func TestDockerContainerCopyFileFromContainer(t *testing.T) {
	fileContent, err := os.ReadFile(filepath.Join(".", "testdata", "hello.sh"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:        nginxImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		Started:      true,
	})
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	copiedFileName := "hello_copy.sh"
	err = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "hello.sh"), "/"+copiedFileName, 700)
	require.NoError(t, err)

	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	require.NoError(t, err)
	require.Equal(t, 0, c, "File %s should exist, expected return code 0, got %v", copiedFileName, c)

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	require.NoError(t, err)
	defer reader.Close()

	fileContentFromContainer, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, fileContent, fileContentFromContainer)
}

func TestDockerContainerCopyEmptyFileFromContainer(t *testing.T) {
	ctx := context.Background()

	nginxC, err := testcontainers.Run(ctx, testcontainers.Request{
		Image:        nginxImage,
		ExposedPorts: []string{nginxDefaultPort},
		WaitingFor:   wait.ForListeningPort(nginxDefaultPort),
		Started:      true,
	})
	testcontainers.CleanupContainer(t, nginxC)
	require.NoError(t, err)

	copiedFileName := "hello_copy.sh"
	err = nginxC.CopyFileToContainer(ctx, filepath.Join(".", "testdata", "empty.sh"), "/"+copiedFileName, 700)
	require.NoError(t, err)

	c, _, err := nginxC.Exec(ctx, []string{"bash", copiedFileName})
	require.NoError(t, err)
	require.Equal(t, 0, c, "File %s should exist, expected return code 0, got %v", copiedFileName, c)

	reader, err := nginxC.CopyFileFromContainer(ctx, "/"+copiedFileName)
	require.NoError(t, err)
	defer reader.Close()

	fileContentFromContainer, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Empty(t, fileContentFromContainer)
}

// creates a temporary dir in which the files will be extracted. Then it will compare the bytes of each file in the source with the bytes from the copied-from-container file
func assertExtractedFiles(t *testing.T, ctx context.Context, ctr *testcontainers.DockerContainer, hostFilePath string, containerFilePath string) {
	// create all copied files into a temporary dir
	tmpDir := t.TempDir()

	// compare the bytes of each file in the source with the bytes from the copied-from-container file
	srcFiles, err := os.ReadDir(hostFilePath)
	require.NoError(t, err)

	for _, srcFile := range srcFiles {
		if srcFile.IsDir() {
			continue
		}
		srcBytes, err := os.ReadFile(filepath.Join(hostFilePath, srcFile.Name()))
		require.NoError(t, err)

		fp := filepath.Join(containerFilePath, srcFile.Name())
		// copy file by file, as there is a limitation in the Docker client to copy an entiry directory from the container
		// paths for the container files are using Linux path separators
		fd, err := ctr.CopyFileFromContainer(ctx, fp)
		require.NoError(t, err, "Path not found in container: %s", fp)
		defer fd.Close()

		targetPath := filepath.Join(tmpDir, srcFile.Name())
		dst, err := os.Create(targetPath)
		require.NoError(t, err)
		defer dst.Close()

		_, err = io.Copy(dst, fd)
		require.NoError(t, err)

		untarBytes, err := os.ReadFile(targetPath)
		require.NoError(t, err)

		assert.Equal(t, srcBytes, untarBytes)
	}
}
