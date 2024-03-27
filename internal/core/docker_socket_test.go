package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckDefaultDockerSocket(t *testing.T) {
	t.Run("Docker client panics", func(tt *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				tt.Errorf("expected panic")
			}
		}()

		checkDefaultDockerSocket(context.Background(), mockCli{infoErr: errors.New("info should panic")}, "")
	})

	t.Run("Local Docker on Unix", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		socket := "unix:///var/run/docker.sock"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Rootless Docker on Unix: home dir", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		tmpDir := tt.TempDir()

		tt.Setenv("HOME", tmpDir)
		runDir := filepath.Join(tmpDir, ".docker", "run")
		err := createTmpDockerSocket(runDir)
		require.NoError(t, err)

		socket := "unix://" + tmpDir + "/.docker/run/docker.sock"
		expected := tmpDir + "/.docker/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Rootless Docker on Unix: XDG_RUNTIME_DIR", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		tmpDir := tt.TempDir()
		tt.Setenv("XDG_RUNTIME_DIR", tmpDir)
		err := createTmpDockerSocket(tmpDir)
		require.NoError(tt, err)

		socket := "unix://" + tmpDir + "/docker.sock"
		expected := tmpDir + "/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Rootless Docker on Unix: home desktop dir", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		tmpDir := tt.TempDir()

		tt.Setenv("HOME", tmpDir)
		desktopDir := filepath.Join(tmpDir, ".docker", "desktop")
		err := createTmpDockerSocket(desktopDir)
		require.NoError(tt, err)

		socket := "unix://" + tmpDir + "/.docker/desktop/docker.sock"
		expected := tmpDir + "/.docker/desktop/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Rootless Docker on Unix: run dir", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		tmpDir := tt.TempDir()

		homeDir := filepath.Join(tmpDir, "home")
		err := createTmpDir(homeDir)
		require.NoError(tt, err)
		tt.Setenv("HOME", homeDir)

		baseRunDir = tmpDir
		tt.Cleanup(func() {
			baseRunDir = originalBaseRunDir
			os.Setenv("HOME", originalHomeDir)
			os.Setenv("USERPROFILE", originalHomeDir)
			os.Setenv("XDG_RUNTIME_DIR", originalXDGRuntimeDir)
		})

		uid := os.Getuid()
		runDir := filepath.Join(tmpDir, "user", fmt.Sprintf("%d", uid))
		err = createTmpDockerSocket(runDir)
		require.NoError(tt, err)

		socket := "unix://" + runDir + "/docker.sock"
		expected := runDir + "/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Local Docker on Windows", func(tt *testing.T) {
		if !IsWindows() {
			tt.Skip("skipping test on non-Windows")
		}

		tt.Setenv("GOOS", "windows")

		socket := "npipe:////./pipe/docker_engine"
		expected := "//./pipe/docker_engine"

		s := checkDefaultDockerSocket(context.Background(), mockCli{}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Docker Desktop on Unix", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		socket := "unix:///var/run/docker.sock"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Docker Desktop on Windows", func(tt *testing.T) {
		if !IsWindows() {
			tt.Skip("skipping test on non-Windows")
		}

		tt.Setenv("GOOS", "windows")

		socket := "npipe:////./pipe/docker_engine"
		expected := "//var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Unix Docker on Unix", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		socket := "tcp://127.0.0.1:12345"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "linux"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Unix Docker on Windows", func(tt *testing.T) {
		if !IsWindows() {
			tt.Skip("skipping test on non-Windows")
		}

		tt.Setenv("GOOS", "windows")

		socket := "tcp://127.0.0.1:12345"
		expected := "//var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "linux"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Windows Docker on Unix", func(tt *testing.T) {
		if IsWindows() {
			tt.Skip("skipping test on Windows")
		}

		socket := "tcp://127.0.0.1:12345"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "windows"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Windows Docker on Windows", func(tt *testing.T) {
		if !IsWindows() {
			tt.Skip("skipping test on non-Windows")
		}

		tt.Setenv("GOOS", "windows")

		socket := "tcp://127.0.0.1:12345"
		expected := "//./pipe/docker_engine"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "windows"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})
}
