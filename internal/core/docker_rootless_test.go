package core

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	originalBaseRunDir    string
	originalXDGRuntimeDir string
	originalHomeDir       string
)

func init() {
	originalBaseRunDir = baseRunDir
	originalXDGRuntimeDir = os.Getenv("XDG_RUNTIME_DIR")
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	originalHomeDir = home
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

func TestRootlessDockerSocketPathNotSupportedOnWindows(t *testing.T) {
	restoreEnvFn := func() {
		os.Setenv("HOME", originalHomeDir)
		os.Setenv("USERPROFILE", originalHomeDir)
		os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
	}

	t.Cleanup(func() {
		restoreEnvFn()
	})

	t.Setenv("GOOS", "windows")
	socketPath, err := rootlessDockerSocketPath(context.Background())
	require.ErrorIs(t, err, ErrRootlessDockerNotSupportedWindows)
	require.Empty(t, socketPath)
}

func TestRootlessDockerSocketPath(t *testing.T) {
	if IsWindows() {
		t.Skip("Docker Rootless is not supported on Windows")
	}

	restoreEnvFn := func() {
		os.Setenv("HOME", originalHomeDir)
		os.Setenv("USERPROFILE", originalHomeDir)
		os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
	}

	t.Cleanup(func() {
		restoreEnvFn()
	})

	t.Run("XDG_RUNTIME_DIR: ${XDG_RUNTIME_DIR}/docker.sock", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Docker Rootless is not supported on Windows")
		}

		tmpDir := t.TempDir()
		t.Setenv("XDG_RUNTIME_DIR", tmpDir)
		err := createTmpDockerSocket(tmpDir)
		require.NoError(t, err)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, socketPath)
	})

	t.Run("Home run dir: ~/.docker/run/docker.sock", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Docker Rootless is not supported on Windows")
		}

		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")
		t.Cleanup(restoreEnvFn)

		runDir := filepath.Join(tmpDir, ".docker", "run")
		err := createTmpDockerSocket(runDir)
		require.NoError(t, err)
		t.Setenv("HOME", tmpDir)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.NoError(t, err)
		assert.Equal(t, DockerSocketSchema+runDir+"/docker.sock", socketPath)
	})

	t.Run("Home desktop dir: ~/.docker/desktop/docker.sock", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Docker Rootless is not supported on Windows")
		}

		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")
		t.Cleanup(restoreEnvFn)

		desktopDir := filepath.Join(tmpDir, ".docker", "desktop")
		err := createTmpDockerSocket(desktopDir)
		require.NoError(t, err)
		t.Setenv("HOME", tmpDir)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.NoError(t, err)
		assert.Equal(t, DockerSocketSchema+desktopDir+"/docker.sock", socketPath)
	})

	t.Run("Run dir: /run/user/${uid}/docker.sock", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Docker Rootless is not supported on Windows")
		}

		tmpDir := t.TempDir()
		_ = os.Unsetenv("XDG_RUNTIME_DIR")

		homeDir := filepath.Join(tmpDir, "home")
		err := createTmpDir(homeDir)
		require.NoError(t, err)
		t.Setenv("HOME", homeDir)

		baseRunDir = tmpDir
		t.Cleanup(func() {
			baseRunDir = originalBaseRunDir
			restoreEnvFn()
		})

		uid := os.Getuid()
		runDir := filepath.Join(tmpDir, "user", strconv.Itoa(uid))
		err = createTmpDockerSocket(runDir)
		require.NoError(t, err)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.NoError(t, err)
		assert.Equal(t, DockerSocketSchema+runDir+"/docker.sock", socketPath)
	})

	t.Run("Rootless not found", func(t *testing.T) {
		if IsWindows() {
			t.Skip("Docker Rootless is not supported on Windows")
		}

		setupRootlessNotFound(t)

		socketPath, err := rootlessDockerSocketPath(context.Background())
		require.ErrorIs(t, err, ErrRootlessDockerNotFoundXDGRuntimeDir)
		require.Empty(t, socketPath)
	})
}

func setupRootlessNotFound(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		baseRunDir = originalBaseRunDir
		os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
	})

	tmpDir := t.TempDir()

	xdgRuntimeDir := filepath.Join(tmpDir, "xdg-runtime-dir")
	err := createTmpDir(xdgRuntimeDir)
	require.NoError(t, err)
	t.Setenv("XDG_RUNTIME_DIR", xdgRuntimeDir)

	homeDir := filepath.Join(tmpDir, "home")
	err = createTmpDir(homeDir)
	require.NoError(t, err)
	t.Setenv("HOME", homeDir)

	homeRunDir := filepath.Join(homeDir, ".docker", "run")
	err = createTmpDir(homeRunDir)
	require.NoError(t, err)

	baseRunDir = tmpDir
	uid := os.Getuid()
	runDir := filepath.Join(tmpDir, "run", "user", strconv.Itoa(uid))
	err = createTmpDir(runDir)
	require.NoError(t, err)
}
