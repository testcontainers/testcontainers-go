package core

import (
	"context"
	"errors"
	"testing"
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
		socket := "unix:///var/run/docker.sock"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Local Docker on Windows", func(tt *testing.T) {
		tt.Setenv("GOOS", "windows")

		socket := "npipe:////./pipe/docker_engine"
		expected := "//./pipe/docker_engine"

		s := checkDefaultDockerSocket(context.Background(), mockCli{}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Docker Desktop on Unix", func(tt *testing.T) {
		socket := "unix:///var/run/docker.sock"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Docker Desktop on Windows", func(tt *testing.T) {
		tt.Setenv("GOOS", "windows")

		socket := "npipe:////./pipe/docker_engine"
		expected := "//var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OS: "Docker Desktop"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Unix Docker on Unix", func(tt *testing.T) {
		socket := "tcp://127.0.0.1:12345"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "linux"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Unix Docker on Windows", func(tt *testing.T) {
		tt.Setenv("GOOS", "windows")

		socket := "tcp://127.0.0.1:12345"
		expected := "//var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "linux"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Windows Docker on Unix", func(tt *testing.T) {
		socket := "tcp://127.0.0.1:12345"
		expected := "/var/run/docker.sock"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "windows"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})

	t.Run("Remote Windows Docker on Windows", func(tt *testing.T) {
		tt.Setenv("GOOS", "windows")

		socket := "tcp://127.0.0.1:12345"
		expected := "//./pipe/docker_engine"

		s := checkDefaultDockerSocket(context.Background(), mockCli{OSType: "windows"}, socket)
		if s != expected {
			tt.Errorf("expected %s, got %s", expected, s)
		}
	})
}
