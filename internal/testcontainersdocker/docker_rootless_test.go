package testcontainersdocker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var originalBaseRunDir string
var originalXDGRuntimeDir string
var originalHomeDir string

func init() {
	originalBaseRunDir = baseRunDir
	originalXDGRuntimeDir = os.Getenv("XDG_RUNTIME_DIR")
	originalHomeDir = os.Getenv("HOME")
}

func TestFileExists(t *testing.T) {
	type cases struct {
		filepath string
		expected bool
	}

	tests := []cases{
		{
			filepath: "testdata",
			expected: true,
		},
		{
			filepath: "docker_rootless.go",
			expected: true,
		},
		{
			filepath: "foobar.doc",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.filepath, func(t *testing.T) {
			result := fileExists(test.filepath)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRootlessDockerSocketPath(t *testing.T) {
	restoreEnvFn := func() {
		os.Setenv("HOME", originalHomeDir)
		os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
	}

	t.Cleanup(func() {
		restoreEnvFn()
	})

	t.Run("Rootless not supported on Windows", func(t *testing.T) {
		t.Setenv("GOOS", "windows")
		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.ErrorIs(t, err, ErrRootlessDockerNotSupportedWindows)
		assert.Empty(t, socketPath)
	})

	t.Run("XDG_RUNTIME_DIR: ${XDG_RUNTIME_DIR}/docker.sock", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_RUNTIME_DIR", tmpDir)
		err := createTmpDockerSocket(tmpDir)
		require.Nil(t, err)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.Nil(t, err)
		assert.NotEmpty(t, socketPath)
	})

	t.Run("Home run dir: ~/.docker/run/docker.sock", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")
		t.Cleanup(restoreEnvFn)

		runDir := filepath.Join(tmpDir, ".docker", "run")
		err := createTmpDockerSocket(runDir)
		require.Nil(t, err)
		t.Setenv("HOME", tmpDir)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.Nil(t, err)
		assert.Equal(t, "unix://"+runDir+"/docker.sock", socketPath)
	})

	t.Run("Home desktop dir: ~/.docker/desktop/docker.sock", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")
		t.Cleanup(restoreEnvFn)

		desktopDir := filepath.Join(tmpDir, ".docker", "desktop")
		err := createTmpDockerSocket(desktopDir)
		require.Nil(t, err)
		t.Setenv("HOME", tmpDir)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.Nil(t, err)
		assert.Equal(t, "unix://"+desktopDir+"/docker.sock", socketPath)
	})

	t.Run("Run dir: /run/user/${uid}/docker.sock", func(t *testing.T) {
		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")

		homeDir := filepath.Join(tmpDir, "home")
		err := createTmpDir(homeDir)
		require.Nil(t, err)
		t.Setenv("HOME", homeDir)

		baseRunDir = tmpDir
		t.Cleanup(func() {
			baseRunDir = originalBaseRunDir
			restoreEnvFn()
		})

		uid := os.Getuid()
		runDir := filepath.Join(tmpDir, "user", fmt.Sprintf("%d", uid))
		err = createTmpDockerSocket(runDir)
		require.Nil(t, err)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.Nil(t, err)
		assert.Equal(t, "unix://"+runDir+"/docker.sock", socketPath)
	})

	t.Run("Rootless not found", func(t *testing.T) {
		setupRootlessNotFound(t)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		assert.ErrorIs(t, err, ErrRootlessDockerNotFound)
		assert.Empty(t, socketPath)

		// the wrapped error includes all the locations that were checked
		assert.ErrorContains(t, err, ErrRootlessDockerNotFoundXDGRuntimeDir.Error())
		assert.ErrorContains(t, err, ErrRootlessDockerNotFoundHomeRunDir.Error())
		assert.ErrorContains(t, err, ErrRootlessDockerNotFoundHomeDesktopDir.Error())
		assert.ErrorContains(t, err, ErrRootlessDockerNotFoundRunDir.Error())
	})
}

func setupRootlessNotFound(t *testing.T) {
	t.Cleanup(func() {
		baseRunDir = originalBaseRunDir
		os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
	})

	tmpDir := t.TempDir()

	xdgRuntimeDir := filepath.Join(tmpDir, "xdg-runtime-dir")
	err := createTmpDir(xdgRuntimeDir)
	require.Nil(t, err)
	t.Setenv("XDG_RUNTIME_DIR", xdgRuntimeDir)

	homeDir := filepath.Join(tmpDir, "home")
	err = createTmpDir(homeDir)
	require.Nil(t, err)
	t.Setenv("HOME", homeDir)

	homeRunDir := filepath.Join(homeDir, ".docker", "run")
	err = createTmpDir(homeRunDir)
	require.Nil(t, err)

	baseRunDir = tmpDir
	uid := os.Getuid()
	runDir := filepath.Join(tmpDir, "run", "user", fmt.Sprintf("%d", uid))
	err = createTmpDir(runDir)
	require.Nil(t, err)
}
